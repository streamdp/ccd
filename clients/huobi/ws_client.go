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

type huobiWs struct {
	l             *log.Logger
	conn          *websocket.Conn
	subscriptions domain.Subscriptions
	subMu         sync.RWMutex
}

var errReconnect = errors.New("reconnect failed")

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger) (*huobiWs, error) {
	h := &huobiWs{
		l:             l,
		subscriptions: domain.Subscriptions{},
	}
	if err := h.reconnect(ctx); err != nil {
		return nil, err
	}
	h.handleWsMessages(ctx, pipe)

	return h, nil
}

func (h *huobiWs) Subscribe(ctx context.Context, from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var (
		id = time.Now().UnixMilli()
		ch = buildChannelName(from, to)
	)
	if err = h.sendSubscribeMsg(ctx, ch, id); err != nil {
		return
	}
	h.subscriptions[ch] = domain.NewSubscription(from, to, id)

	return
}

func (h *huobiWs) ListSubscriptions() domain.Subscriptions {
	s := make(domain.Subscriptions, len(h.subscriptions))
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	for k, v := range h.subscriptions {
		s[k] = v
	}

	return s
}

func (h *huobiWs) Unsubscribe(ctx context.Context, from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if c, ok := h.subscriptions[ch]; ok {
		if err = h.sendUnsubscribeMsg(ctx, ch, c.Id()); err != nil {
			return
		}
		delete(h.subscriptions, ch)
	}

	return
}

func (h *huobiWs) reconnect(ctx context.Context) error {
	if h.conn != nil {
		if err := h.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			h.l.Println(err)
			// reducing logs and CPU load when API key expired
			time.Sleep(10 * time.Second)
		}
	}

	var err error
	if h.conn, _, err = websocket.Dial(ctx, wssUrl, nil); err != nil {
		return fmt.Errorf("failed to dial ws server: %w", err)
	}

	return nil
}

func (h *huobiWs) resubscribe(ctx context.Context) (err error) {
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	for k, v := range h.subscriptions {
		if err = h.sendSubscribeMsg(ctx, k, v.Id()); err != nil {
			return
		}
	}

	return
}

func (h *huobiWs) handleWsError(ctx context.Context, err error) error {
	h.l.Println(err)
	for {
		select {
		case <-time.After(time.Minute):
			return errReconnect
		default:
			if err = h.reconnect(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}
			if err = h.resubscribe(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}

			return nil
		}
	}
}

func (h *huobiWs) handleWsMessages(ctx context.Context, pipe chan *domain.Data) {
	go func() {
		defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
			if errClose := conn.Close(code, reason); errClose != nil {
				h.l.Println(errClose)
			}
		}(h.conn, websocket.StatusNormalClosure, "")
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
				if _, r, err = h.conn.Reader(ctx); err != nil {
					if err = h.handleWsError(ctx, err); err != nil {
						h.l.Println(err)

						return
					}

					continue
				}
				if body, err = gzipDecompress(r); err != nil {
					h.l.Println(err)

					continue
				}
				if bytes.Contains(body, []byte("ping")) {
					if err = h.pingHandler(ctx, body); err != nil {
						if err = h.handleWsError(ctx, err); err != nil {
							h.l.Println(err)

							return
						}
					}

					continue
				}
				data := &huobiWsData{}
				if err = json.Unmarshal(body, data); err != nil {
					h.l.Println(err)

					continue
				}
				if data.Ch == "" {
					continue
				}
				from, to := h.pairFromChannelName(data.Ch)
				if from != "" && to != "" {
					pipe <- convertHuobiWsDataToDomain(from, to, data)
				}
			}
		}
	}()
}

func (h *huobiWs) pingHandler(ctx context.Context, m []byte) error {
	m = bytes.ReplaceAll(m, []byte("ping"), []byte("pong"))
	if err := h.conn.Write(ctx, websocket.MessageText, m); err != nil {
		return fmt.Errorf("failed to send pong response: %w", err)
	}

	return nil
}

func (h *huobiWs) pairFromChannelName(ch string) (string, string) {
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	if c, ok := h.subscriptions[ch]; ok {
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

func (h *huobiWs) sendUnsubscribeMsg(ctx context.Context, ch string, id int64) error {
	if err := h.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"unsub\": \"%s\", \"id\":\"%d\"}", ch, id)),
	); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

func (h *huobiWs) sendSubscribeMsg(ctx context.Context, ch string, id int64) error {
	if err := h.conn.Write(ctx, websocket.MessageText, []byte(
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

func convertHuobiWsDataToDomain(from, to string, d *huobiWsData) *domain.Data {
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
