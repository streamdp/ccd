package wsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/domain"
)

const (
	defaultReconnectTimeout = 10 * time.Second
	defaultPingTimeout      = 5 * time.Second
)

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

	sessionRepo clients.SessionRepo
}

var (
	ErrReconnect                  = errors.New("reconnect failed")
	ErrClientReconnected          = errors.New("ws client has been reconnected")
	ErrWsConnectionNotInitialized = errors.New("ws connection is not initialized")
)

func New(ctx context.Context, wsUrl string, sessionRepo clients.SessionRepo, l *log.Logger) *Ws {
	w := &Ws{
		l: l,

		wsUrl: wsUrl,

		subscriptions: domain.Subscriptions{},

		up:   make(chan struct{}),
		down: make(chan struct{}),

		sessionRepo: sessionRepo,
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

	if err = w.sessionRepo.AddTask(ctx, buildWsSessionName(from, to), 0); err != nil {
		w.l.Println("failed to add subscription to the session repo: " + err.Error())
	}

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

	if err := w.sessionRepo.RemoveTask(ctx, buildWsSessionName(from, to)); err != nil {
		w.l.Println("failed to add subscription to the session repo: " + err.Error())
	}

	w.WsDown()

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
	if w.conn == nil {
		return nil, ErrWsConnectionNotInitialized
	}

	_, body, err := w.conn.Read(ctx)
	if err != nil {
		if errors.As(err, &websocket.CloseError{Code: websocket.StatusNormalClosure}) ||
			errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("failed to ws read: %w", err)
		}

		if err = w.HandleWsError(ctx, err); err != nil {
			return nil, fmt.Errorf("failed to handle ws client error: %w", err)
		}

		return nil, ErrClientReconnected
	}

	return body, nil
}

func (w *Ws) Reader(ctx context.Context) (io.Reader, error) {
	if w.conn == nil {
		return nil, ErrWsConnectionNotInitialized
	}

	_, r, err := w.conn.Reader(ctx)
	if err != nil {
		if errors.As(err, &websocket.CloseError{Code: websocket.StatusNormalClosure}) ||
			errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("failed to get ws reader: %w", err)
		}

		if err = w.HandleWsError(ctx, err); err != nil {
			return nil, fmt.Errorf("failed to handle ws client error: %w", err)
		}

		return nil, ErrClientReconnected
	}

	return r, nil
}

func (w *Ws) RestoreLastSession(ctx context.Context) error {
	if w.sessionRepo == nil {
		return nil
	}

	sessions, err := w.sessionRepo.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	for session := range sessions {
		if pair := strings.Split(session, ":"); len(pair) == 3 {
			if err = w.Subscribe(ctx, pair[1], pair[2]); err != nil {
				return fmt.Errorf("failed to restore last ws session: %w", err)
			}
		}
	}

	return nil
}

func (w *Ws) WsDown() {
	w.down <- struct{}{}
	<-w.down
}

func (w *Ws) wsUp() {
	w.up <- struct{}{}
	<-w.up
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
		cancel context.CancelFunc = func() {}

		isConnected bool
	)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			cancel()

			return
		case <-w.up:
			if !isConnected || w.isConnectionBroken(ctx) {
				cancel()

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
			if isConnected && len(w.subscriptions) == 0 || w.isConnectionBroken(ctx) {
				cancel()

				if err := w.conn.Close(websocket.StatusNormalClosure, "no active subscriptions"); err != nil {
					w.l.Printf("failed to close websocket connection: %v", err)
				}
				w.conn = nil

				w.subMu.Lock()
				w.subscriptions = domain.Subscriptions{}
				w.subMu.Unlock()

				w.l.Printf("websocket connection closed (no active subscriptions were found)")
				isConnected = false
			}
			w.down <- struct{}{}
		}
	}
}

func (w *Ws) isConnectionBroken(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	return w.conn != nil && w.conn.Ping(ctx) != nil
}

func buildWsSessionName(from, to string) string {
	return fmt.Sprintf("WS:%s:%s", from, to)
}
