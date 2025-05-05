package kraken

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/domain"
)

const wssUrl = "wss://ws.kraken.com/v2"

type ws struct {
	l             *log.Logger
	conn          *websocket.Conn
	subscriptions domain.Subscriptions
	subMu         sync.RWMutex
}

var errReconnect = errors.New("reconnect failed")

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger) (*ws, error) {
	k := &ws{
		l:             l,
		subscriptions: domain.Subscriptions{},
	}
	if err := k.reconnect(ctx); err != nil {
		return nil, err
	}
	go k.handleWsMessages(ctx, pipe)

	return k, nil
}

func (w *ws) Subscribe(ctx context.Context, from, to string) error {
	var (
		id = time.Now().UnixMilli()
		ch = buildChannelName(from, to)
	)
	w.l.Printf("trying to subscribe on the %s/%s pair", from, to)
	if err := w.sendSubscribeMsg(ctx, ch); err != nil {
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
	defer w.subMu.RUnlock()
	for k, v := range w.subscriptions {
		s[k] = v
	}

	return s
}

func (w *ws) Unsubscribe(ctx context.Context, from, to string) error {
	w.subMu.Lock()
	defer w.subMu.Unlock()
	ch := buildChannelName(from, to)
	if _, ok := w.subscriptions[ch]; ok {
		w.l.Printf("trying to unsubscribe from the %s/%s pair", from, to)
		if err := w.sendUnsubscribeMsg(ctx, ch); err != nil {
			return fmt.Errorf("failed to wss unsubscribes: %w", err)
		}
		delete(w.subscriptions, ch)
	}

	return nil
}

func (w *ws) reconnect(ctx context.Context) error {
	if w.conn != nil {
		if err := w.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
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
	for k := range w.subscriptions {
		if err := w.sendSubscribeMsg(ctx, k); err != nil {
			return fmt.Errorf("failed to wss resubscribe: %w", err)
		}
	}

	return nil
}

func (w *ws) handleWsError(ctx context.Context, err error) error {
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

	switch msg.Method {
	case "subscribe":
		if msg.Error != "" {
			return "failed to subscribe: " + msg.Error
		}
		if msg.Success {
			return fmt.Sprintf(
				"%s channel: successfully subscribed on the %s pair", msg.Result.Channel, msg.Result.Symbol)
		}
	case "unsubscribe":
		if msg.Error != "" {
			return "failed to unsubscribe: " + msg.Error
		}
		if msg.Success {
			return fmt.Sprintf(
				"%s channel: successfully unsubscribed from the %s pair", msg.Result.Channel, msg.Result.Symbol)
		}
	}

	return ""
}

func (w *ws) handleWsUpdate(body []byte, pipe chan *domain.Data) error {
	data := &wsData{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("failed to unmarshal ws update message: %w", err)
	}

	if data.Channel != "ticker" || len(data.Data) == 0 {
		return nil
	}

	for _, tick := range data.Data {
		from, to := w.pairFromChannelName(tick.Symbol)
		if from == "" || to == "" {
			continue
		}

		pipe <- convertWsDataToDomain(from, to, &tick, time.Now().UTC().UnixMilli())
	}

	return nil
}

func (w *ws) handleWsMessages(ctx context.Context, pipe chan *domain.Data) {
	defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
		if errClose := conn.Close(code, reason); errClose != nil {
			w.l.Println(errClose)
		}
	}(w.conn, websocket.StatusNormalClosure, "")
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
			if body, err = io.ReadAll(r); err != nil {
				w.l.Println(err)

				continue
			}

			if bytes.Contains(body, []byte("method")) {
				if msg := w.handleServerResponse(body); msg != "" {
					w.l.Println(msg)
				}

				continue
			}

			if bytes.Contains(body, []byte("heartbeat")) {
				continue
			}

			if err = w.handleWsUpdate(body, pipe); err != nil {
				w.l.Println(err)
			}
		}
	}
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
	return fmt.Sprintf("%s/%s", from, to)
}

func (w *ws) sendUnsubscribeMsg(ctx context.Context, ch string) error {
	msg, err := json.Marshal(wsMessage{
		Method: "unsubscribe",
		Params: &wsMessageParams{
			Channel: "ticker",
			Symbol:  []string{ch},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe message: %w", err)
	}

	if err = w.conn.Write(ctx, websocket.MessageText, msg); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

func (w *ws) sendSubscribeMsg(ctx context.Context, ch string) error {
	msg, err := json.Marshal(wsMessage{
		Method: "subscribe",
		Params: &wsMessageParams{
			Channel: "ticker",
			Symbol:  []string{ch},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal subscribe message: %w", err)
	}

	if err = w.conn.Write(ctx, websocket.MessageText, msg); err != nil {
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	return nil
}

func convertWsDataToDomain(from, to string, tick *wsTickerInfo, lastUpdate int64) *domain.Data {
	if tick == nil {
		return nil
	}

	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:      from,
		ToSymbol:        to,
		Change24Hour:    tick.Change,
		ChangePct24Hour: tick.ChangePct,
		Volume24Hour:    tick.Volume,
		High24Hour:      tick.High,
		Low24Hour:       tick.Low,
		Price:           tick.Vwap,
		LastUpdate:      lastUpdate,
	})

	return &domain.Data{
		FromSymbol:      from,
		ToSymbol:        to,
		Change24Hour:    tick.Change,
		ChangePct24Hour: tick.ChangePct,
		Volume24Hour:    tick.Volume,
		High24Hour:      tick.High,
		Low24Hour:       tick.Low,
		Price:           tick.Vwap,
		LastUpdate:      lastUpdate,
		DisplayDataRaw:  string(b),
	}
}
