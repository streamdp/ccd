package cryptocompare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
)

const wssUrl = "wss://streamer.cryptocompare.com/v2"

type ws struct {
	l             *log.Logger
	conn          *websocket.Conn
	apiKey        string
	subscriptions domain.Subscriptions
	subMu         sync.RWMutex
}

var (
	errReconnect = errors.New("reconnect failed")
	errHeartbeat = errors.New("heartbeat loss")
)

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger, cfg *config.App) (*ws, error) {
	if cfg.ApiKey == "" {
		return nil, errApiKeyNotDefined
	}

	h := &ws{
		l:             l,
		apiKey:        cfg.ApiKey,
		subscriptions: domain.Subscriptions{},
	}
	if err := h.reconnect(ctx); err != nil {
		return nil, err
	}
	go h.handleWsMessages(ctx, pipe)

	return h, nil
}

func (w *ws) Subscribe(ctx context.Context, from, to string) error {
	var ch = buildChannelName(from, to)
	if err := w.sendSubscribeMsg(ctx, ch); err != nil {
		return fmt.Errorf("failed to wss subscribe: %w", err)
	}

	w.subMu.Lock()
	w.subscriptions[ch] = domain.NewSubscription(from, to, 0)
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
	var ch = buildChannelName(from, to)
	if _, ok := w.subscriptions[ch]; ok {
		if err := w.sendUnsubscribeMsg(ctx, ch); err != nil {
			return fmt.Errorf("failed to wss unsubscribe: %w", err)
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
	u, err := w.buildURL()
	if err != nil {
		return err
	}
	if w.conn, _, err = websocket.Dial(ctx, u.String(), nil); err != nil {
		return fmt.Errorf("failed to dial ws server: %w", err)
	}

	return nil
}

func (w *ws) buildURL() (*url.URL, error) {
	u, err := url.Parse(wssUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	q := u.Query()
	q.Set("api_key", w.apiKey)
	u.RawQuery = q.Encode()

	return u, nil
}

func buildChannelName(from, to string) string {
	return fmt.Sprintf("5~CCCAGG~%s~%s", strings.ToUpper(from), strings.ToUpper(to))
}

func (w *ws) sendSubscribeMsg(ctx context.Context, ch string) error {
	if err := w.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubAdd\",\"subs\":[\"%s\"]}", ch)),
	); err != nil {
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	return nil
}

func (w *ws) sendUnsubscribeMsg(ctx context.Context, ch string) error {
	if err := w.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubRemove\",\"subs\":[\"%s\"]}", ch)),
	); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

func (w *ws) resubscribe(ctx context.Context) error {
	w.subMu.RLock()
	defer w.subMu.RUnlock()
	for k := range w.subscriptions {
		if err := w.sendSubscribeMsg(ctx, k); err != nil {
			return fmt.Errorf("failed to resubscribe: %w", err)
		}
	}

	return nil
}

func (w *ws) handleWssError(ctx context.Context, err error) error {
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

func (w *ws) handleWsMessages(ctx context.Context, pipe chan *domain.Data) {
	defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
		if errClose := conn.Close(code, reason); errClose != nil {
			w.l.Println(errClose)
		}
	}(w.conn, websocket.StatusNormalClosure, "")
	var (
		hb      = newHeartbeat()
		hbTimer = time.NewTimer(heartbeatCheckInterval)
	)
	defer hbTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-hbTimer.C:
			hbTimer.Reset(heartbeatCheckInterval)
			if hb.isLost() {
				if err := w.handleWssError(ctx, errHeartbeat); err != nil {
					w.l.Println(err)

					return
				}
			}
			hb.decrease()
		default:
			var (
				body []byte
				err  error
			)
			if _, body, err = w.conn.Read(ctx); err != nil {
				if err = w.handleWssError(ctx, err); err != nil {
					w.l.Println(err)

					return
				}

				continue
			}
			data := &wsData{}
			if err = json.Unmarshal(body, data); err != nil {
				w.l.Println(err)

				continue
			}
			switch data.Type {
			case "999":
				hb.reset()
			case "5":
				pipe <- convertWsDataToDomain(data)
			}
		}
	}
}

func convertWsDataToDomain(d *wsData) *domain.Data {
	if d == nil {
		return nil
	}
	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:     d.FromSymbol,
		ToSymbol:       d.ToSymbol,
		Open24Hour:     d.Open24Hour,
		Volume24Hour:   d.Volume24Hour,
		Volume24HourTo: d.Volume24HourTo,
		High24Hour:     d.High24Hour,
		Price:          d.Price,
		LastUpdate:     d.LastUpdate,
		Supply:         d.CurrentSupply,
		MktCap:         d.CurrentSupplyMktCap,
	})

	return &domain.Data{
		FromSymbol:     d.FromSymbol,
		ToSymbol:       d.ToSymbol,
		Open24Hour:     d.Open24Hour,
		Volume24Hour:   d.Volume24Hour,
		Low24Hour:      d.Low24Hour,
		High24Hour:     d.High24Hour,
		Price:          d.Price,
		Supply:         d.CurrentSupply,
		MktCap:         d.CurrentSupplyMktCap,
		LastUpdate:     d.LastUpdate,
		DisplayDataRaw: string(b),
	}
}
