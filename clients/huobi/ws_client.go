package huobi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/domain"
)

const wssUrl = "wss://api.huobi.pro/ws"

type ws struct {
	l    *log.Logger
	conn *websocket.Conn

	subscriptions domain.Subscriptions
	subMu         sync.RWMutex

	pipe chan *domain.Data

	up   chan struct{}
	down chan struct{}
}

var errReconnect = errors.New("reconnect failed")

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger) (*ws, error) {
	w := &ws{
		l:             l,
		subscriptions: domain.Subscriptions{},

		pipe: pipe,

		up:   make(chan struct{}),
		down: make(chan struct{}),
	}

	go w.serveWsConnection(ctx)

	return w, nil
}

func (w *ws) Subscribe(ctx context.Context, from, to string) error {
	w.wsUp()

	var (
		id = time.Now().UnixMilli()
		ch = buildChannelName(from, to)
	)
	if err := w.sendSubscribeMsg(ctx, ch, id); err != nil {
		return fmt.Errorf("failed to ws subscribe: %w", err)
	}

	w.subMu.Lock()
	w.subscriptions[ch] = domain.NewSubscription(from, to, id)
	w.subMu.Unlock()

	return nil
}

func (w *ws) ListSubscriptions() domain.Subscriptions {
	s := make(domain.Subscriptions, len(w.subscriptions))
	w.subMu.RLock()
	for k, v := range w.subscriptions {
		s[k] = v
	}
	w.subMu.RUnlock()

	return s
}

func (w *ws) Unsubscribe(ctx context.Context, from, to string) error {
	w.subMu.Lock()
	defer w.subMu.Unlock()

	var ch = buildChannelName(from, to)
	if c, ok := w.subscriptions[ch]; ok {
		if err := w.sendUnsubscribeMsg(ctx, ch, c.Id()); err != nil {
			return fmt.Errorf("failed to wss unsubscribes: %w", err)
		}
		delete(w.subscriptions, ch)
	}

	w.wsDown()

	return nil
}

func (w *ws) serveWsConnection(ctx context.Context) {
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
			if !isConnected {
				ctxUp, cancel = context.WithCancel(ctx)
				if err := w.reconnect(ctxUp); err != nil {
					w.l.Println(err)

					continue
				}
				go w.handleWsMessages(ctxUp, w.pipe)

				w.l.Printf("websocket connection open")
				isConnected = true
			}
			w.up <- struct{}{}
		case <-w.down:
			if isConnected && len(w.subscriptions) == 0 {
				if cancel != nil {
					cancel()
				}

				if err := w.conn.CloseNow(); err != nil {
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

func (w *ws) wsUp() {
	w.up <- struct{}{}
	<-w.up
}

func (w *ws) wsDown() {
	w.down <- struct{}{}
	<-w.down
}

func (w *ws) reconnect(ctx context.Context) error {
	if w.conn != nil {
		if err := w.conn.CloseNow(); !errors.As(err, &websocket.CloseError{}) && !errors.Is(err, context.Canceled) {
			w.l.Println(err)
			// reducing logs and CPU load when API key expired
			time.Sleep(10 * time.Second)
		}
	}

	var err error
	if w.conn, _, err = websocket.Dial(ctx, wssUrl, nil); err != nil {
		return fmt.Errorf("failed to dial ws server: %w", err)
	}

	return nil
}

func (w *ws) resubscribe(ctx context.Context) error {
	w.subMu.RLock()
	defer w.subMu.RUnlock()

	for k, v := range w.subscriptions {
		if err := w.sendSubscribeMsg(ctx, k, v.Id()); err != nil {
			return fmt.Errorf("failed to wss resubscribe: %w", err)
		}
	}

	return nil
}

func (w *ws) handleWsError(ctx context.Context, err error) error {
	if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.Canceled) {
		return nil
	}

	w.l.Println(err)
	for {
		select {
		case <-time.After(time.Minute):
			return errReconnect
		default:
			if err = w.reconnect(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}
			if err = w.resubscribe(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}

			return nil
		}
	}
}

func (w *ws) handleServerResponse(body []byte) string {
	msg := &wsMessage{}
	if err := json.Unmarshal(body, msg); err != nil {
		return "failed to unmarshal server response: " + err.Error()
	}

	if msg.Status == "ok" {
		if msg.Subbed != "" {
			return "ticker channel: successfully subscribed on the " + msg.Subbed
		}
		if msg.Unsubbed != "" {
			return "ticker channel: successfully unsubscribed from the " + msg.Unsubbed
		}
	} else {
		return "ticker channel: failed to sub/unsub operation"
	}

	return ""
}

func (w *ws) handleWsUpdate(body []byte, pipe chan *domain.Data) error {
	data := &wsData{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("failed to unmarshal ws update message: %w", err)
	}

	if data.Ch == "" {
		return nil
	}

	from, to := w.pairFromChannelName(data.Ch)
	if from == "" || to == "" {
		return nil
	}

	pipe <- convertWsDataToDomain(from, to, data)

	return nil
}

func (w *ws) handleWsMessages(ctx context.Context, pipe chan *domain.Data) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var (
				r    io.Reader
				body []byte
				err  error
			)
			if _, r, err = w.conn.Reader(ctx); err != nil {
				if err = w.handleWsError(ctx, err); err != nil {
					w.l.Println(err)

					return
				}

				continue
			}
			if body, err = gzipDecompress(r); err != nil {
				w.l.Println(err)

				continue
			}
			if bytes.Contains(body, []byte("ping")) {
				if err = w.pingHandler(ctx, body); err != nil {
					w.l.Println(err)
				}

				continue
			}
			if bytes.Contains(body, []byte("subbed")) {
				if msg := w.handleServerResponse(body); msg != "" {
					w.l.Println(msg)
				}

				continue
			}
			if err = w.handleWsUpdate(body, pipe); err != nil {
				w.l.Println(err)
			}
		}
	}
}

func (w *ws) pingHandler(ctx context.Context, m []byte) error {
	m = bytes.ReplaceAll(m, []byte("ping"), []byte("pong"))
	if err := w.conn.Write(ctx, websocket.MessageText, m); err != nil {
		return fmt.Errorf("failed to send pong response: %w", err)
	}

	return nil
}

func (w *ws) pairFromChannelName(ch string) (string, string) {
	w.subMu.RLock()
	defer w.subMu.RUnlock()

	if c, ok := w.subscriptions[ch]; ok {
		return c.From, c.To
	}

	return "", ""
}

func buildChannelName(from, to string) string {
	if strings.ToLower(to) == "usd" {
		to = "usdt"
	}

	return fmt.Sprintf("market.%s.ticker", strings.ToLower(from+to))
}

func (w *ws) sendUnsubscribeMsg(ctx context.Context, ch string, id int64) error {
	if err := w.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"unsub\": \"%s\", \"id\":\"%d\"}", ch, id)),
	); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

func (w *ws) sendSubscribeMsg(ctx context.Context, ch string, id int64) error {
	if err := w.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"sub\": \"%s\", \"id\":\"%d\"}", ch, id)),
	); err != nil {
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	return nil
}

func gzipDecompress(r io.Reader) ([]byte, error) {
	r, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read uncompressed data: %w", err)
	}

	return data, nil
}

func convertWsDataToDomain(from, to string, d *wsData) *domain.Data {
	if d == nil {
		return nil
	}
	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     d.Tick.Open,
		Volume24Hour:   d.Tick.Amount,
		Volume24HourTo: d.Tick.Vol,
		High24Hour:     d.Tick.High,
		Price:          d.Tick.Bid,
		LastUpdate:     d.Ts,
		Supply:         float64(d.Tick.Count),
	})

	return &domain.Data{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     d.Tick.Open,
		Volume24Hour:   d.Tick.Amount,
		Low24Hour:      d.Tick.Low,
		High24Hour:     d.Tick.High,
		Price:          d.Tick.Bid,
		Supply:         float64(d.Tick.Count),
		LastUpdate:     d.Ts,
		DisplayDataRaw: string(b),
	}
}
