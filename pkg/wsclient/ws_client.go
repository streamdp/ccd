package wsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/domain"
)

const defaultReconnectTimeout = 10 * time.Second

type Ws struct {
	l    *log.Logger
	conn *websocket.Conn

	wsUrl string

	subscriptions domain.Subscriptions
	subMu         sync.RWMutex

	ChannelNameBuilder        func(from, to string) string
	SubscribeMessageBuilder   func(ch string, id int64) ([]byte, error)
	UnsubscribeMessageBuilder func(ch string, id int64) ([]byte, error)
	PingMessageBuilder        func(ch string, id int64) ([]byte, error)
	PongMessageBuilder        func(ch string, id int64) ([]byte, error)

	MessageHandler func(ctx context.Context)

	up   chan struct{}
	down chan struct{}
}

var (
	ErrReconnect                  = errors.New("reconnect failed")
	ErrWsConnectionNotInitialized = errors.New("ws connection is not initialized")
)

func New(ctx context.Context, wsUrl string, l *log.Logger) *Ws {
	w := &Ws{
		l: l,

		wsUrl: wsUrl,

		subscriptions: domain.Subscriptions{},

		up:   make(chan struct{}),
		down: make(chan struct{}),
	}

	go w.serveWsConnection(ctx)

	return w
}

func (w *Ws) Subscribe(ctx context.Context, from, to string) error {
	w.wsUp()

	id := time.Now().UnixMilli()
	ch := w.ChannelNameBuilder(from, to)

	msg, err := w.SubscribeMessageBuilder(ch, id)
	if err != nil {
		return fmt.Errorf("failed to build subscribe message: %w", err)
	}

	if err = w.sendMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to ws subscribe: %w", err)
	}

	w.subMu.Lock()
	w.subscriptions[ch] = domain.NewSubscription(from, to, id)
	w.subMu.Unlock()

	return nil
}

func (w *Ws) ListSubscriptions() domain.Subscriptions {
	s := make(domain.Subscriptions, len(w.subscriptions))
	w.subMu.RLock()
	for k, v := range w.subscriptions {
		s[k] = v
	}
	w.subMu.RUnlock()

	return s
}

func (w *Ws) Unsubscribe(ctx context.Context, from, to string) error {
	ch := w.ChannelNameBuilder(from, to)
	w.subMu.RLock()
	sub, ok := w.subscriptions[ch]
	w.subMu.RUnlock()

	if ok {
		msg, err := w.UnsubscribeMessageBuilder(ch, sub.Id())
		if err != nil {
			return fmt.Errorf("failed to build subscribe message: %w", err)
		}

		if err = w.sendMessage(ctx, msg); err != nil {
			return fmt.Errorf("failed to ws unsubscribe: %w", err)
		}

		w.subMu.Lock()
		delete(w.subscriptions, ch)
		w.subMu.Unlock()
	}

	w.wsDown()

	return nil
}

func (w *Ws) HandleWsError(ctx context.Context, err error) error {
	w.l.Println(err)
	for {
		select {
		case <-time.After(time.Minute):
			return ErrReconnect
		default:
			if err = w.reconnect(ctx); err != nil {
				continue
			}
			if err = w.resubscribe(ctx); err != nil {
				continue
			}

			return nil
		}
	}
}

func (w *Ws) PairFromChannelName(ch string) (string, string) {
	w.subMu.RLock()
	defer w.subMu.RUnlock()

	if c, ok := w.subscriptions[ch]; ok {
		return c.From, c.To
	}

	return "", ""
}

func (w *Ws) Ping(ctx context.Context, ch string, id int64) error {
	msg, err := w.PingMessageBuilder(ch, id)
	if err != nil {
		return fmt.Errorf("failed to build ping message: %w", err)
	}

	if err = w.sendMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to ping ws: %w", err)
	}

	return nil
}

func (w *Ws) Pong(ctx context.Context, ch string, id int64) error {
	msg, err := w.PongMessageBuilder(ch, id)
	if err != nil {
		return fmt.Errorf("failed to build pong message: %w", err)
	}

	if err = w.sendMessage(ctx, msg); err != nil {
		return fmt.Errorf("failed to pong ws: %w", err)
	}

	return nil
}

func (w *Ws) Read(ctx context.Context) ([]byte, error) {
	_, body, err := w.conn.Read(ctx)
	if err != nil {
		if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if err = w.HandleWsError(ctx, err); err != nil {
			return nil, err
		}
	}

	return body, nil
}

func (w *Ws) Reader(ctx context.Context) (io.Reader, error) {
	_, r, err := w.conn.Reader(ctx)
	if err != nil {
		if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.Canceled) {
			return nil, err
		}

		if err = w.HandleWsError(ctx, err); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (w *Ws) reconnect(ctx context.Context) error {
	var err error

	if w.conn != nil {
		err = w.conn.Close(websocket.StatusNormalClosure, "")
		if !errors.As(err, &websocket.CloseError{}) && !errors.Is(err, context.Canceled) {
			w.l.Println(err)

			time.Sleep(defaultReconnectTimeout)
		}
		w.conn = nil
	}

	if w.conn, _, err = websocket.Dial(ctx, w.wsUrl, nil); err != nil {
		return fmt.Errorf("failed to dial ws server: %w", err)
	}

	return nil
}
func (w *Ws) resubscribe(ctx context.Context) error {
	w.subMu.RLock()
	defer w.subMu.RUnlock()

	for ch, sub := range w.subscriptions {
		msg, err := w.SubscribeMessageBuilder(ch, sub.Id())
		if err != nil {
			return fmt.Errorf("failed to build subscribe message: %w", err)
		}
		if err = w.sendMessage(ctx, msg); err != nil {
			return fmt.Errorf("failed to resubscribe: %w", err)
		}
	}

	return nil
}
func (w *Ws) sendMessage(ctx context.Context, message []byte) error {
	if w.conn == nil {
		return ErrWsConnectionNotInitialized
	}

	if err := w.conn.Write(ctx, websocket.MessageText, message); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (w *Ws) serveWsConnection(ctx context.Context) {
	defer close(w.up)
	defer close(w.down)

	var (
		ctxUp  context.Context
		cancel context.CancelFunc

		isConnected bool
	)
	defer func() {
		if cancel != nil {
			cancel()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}

			return
		case <-w.up:
			isConnectionAlive := w.conn != nil && w.conn.Ping(ctx) == nil
			if !isConnected || !isConnectionAlive {
				ctxUp, cancel = context.WithCancel(ctx)
				if err := w.reconnect(ctxUp); err != nil {
					w.l.Println(err)
					w.up <- struct{}{}

					continue
				}
				go w.MessageHandler(ctxUp)

				w.l.Printf("websocket connection open")
				isConnected = true
			}
			w.up <- struct{}{}
		case <-w.down:
			isConnectionBroken := w.conn != nil && w.conn.Ping(ctx) != nil
			if isConnected && len(w.subscriptions) == 0 || isConnectionBroken {
				if cancel != nil {
					cancel()
				}

				if err := w.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
					w.l.Printf("failed to close websocket connection: %v", err)
				}
				w.conn = nil

				w.l.Printf("websocket connection closed (no active subscriptions were found)")
				isConnected = false
			}
			w.down <- struct{}{}
		}
	}
}

func (w *Ws) wsUp() {
	w.up <- struct{}{}
	<-w.up
}

func (w *Ws) wsDown() {
	w.down <- struct{}{}
	<-w.down
}
