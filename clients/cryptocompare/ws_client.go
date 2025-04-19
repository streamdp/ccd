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
	"github.com/streamdp/ccd/domain"
)

const wssUrl = "wss://streamer.cryptocompare.com/v2"

type cryptoCompareWs struct {
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

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger) (*cryptoCompareWs, error) {
	apiKey, err := getApiKey()
	if err != nil {
		return nil, err
	}
	h := &cryptoCompareWs{
		l:             l,
		apiKey:        apiKey,
		subscriptions: domain.Subscriptions{},
	}
	if err = h.reconnect(ctx); err != nil {
		return nil, err
	}
	h.handleWsMessages(ctx, pipe)

	return h, nil
}

func (c *cryptoCompareWs) Subscribe(ctx context.Context, from, to string) (err error) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if err = c.sendSubscribeMsg(ctx, ch); err != nil {
		return
	}
	c.subscriptions[ch] = domain.NewSubscription(from, to, 0)

	return
}

func (c *cryptoCompareWs) ListSubscriptions() domain.Subscriptions {
	s := make(domain.Subscriptions, len(c.subscriptions))
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	for k, v := range c.subscriptions {
		s[k] = v
	}

	return s
}

func (c *cryptoCompareWs) Unsubscribe(ctx context.Context, from, to string) error {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if _, ok := c.subscriptions[ch]; ok {
		if err := c.sendUnsubscribeMsg(ctx, ch); err != nil {
			return err
		}
		delete(c.subscriptions, ch)
	}

	return nil
}

func (c *cryptoCompareWs) reconnect(ctx context.Context) error {
	if c.conn != nil {
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			c.l.Println(err)
			// reducing logs and CPU load when API key expired
			time.Sleep(10 * time.Second)
		}
	}
	u, err := c.buildURL()
	if err != nil {
		return err
	}
	if c.conn, _, err = websocket.Dial(ctx, u.String(), nil); err != nil {
		return fmt.Errorf("failed to dial ws server: %w", err)
	}

	return nil
}

func (c *cryptoCompareWs) buildURL() (*url.URL, error) {
	u, err := url.Parse(wssUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	q := u.Query()
	q.Set("api_key", c.apiKey)
	u.RawQuery = q.Encode()

	return u, nil
}

func buildChannelName(from, to string) string {
	return fmt.Sprintf("5~CCCAGG~%s~%s", strings.ToUpper(from), strings.ToUpper(to))
}

func (c *cryptoCompareWs) sendSubscribeMsg(ctx context.Context, ch string) error {
	if err := c.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubAdd\",\"subs\":[\"%s\"]}", ch)),
	); err != nil {
		return fmt.Errorf("failed to send subscribe message: %w", err)
	}

	return nil
}

func (c *cryptoCompareWs) sendUnsubscribeMsg(ctx context.Context, ch string) error {
	if err := c.conn.Write(ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubRemove\",\"subs\":[\"%s\"]}", ch)),
	); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %w", err)
	}

	return nil
}

func (c *cryptoCompareWs) resubscribe(ctx context.Context) error {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	for k := range c.subscriptions {
		if err := c.sendSubscribeMsg(ctx, k); err != nil {
			return fmt.Errorf("failed to resubscribe: %w", err)
		}
	}

	return nil
}

func (c *cryptoCompareWs) handleWssError(ctx context.Context, err error) error {
	c.l.Println(err)
	for {
		select {
		case <-time.After(time.Minute):
			return errReconnect
		default:
			if err = c.reconnect(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}
			if err = c.resubscribe(ctx); err != nil {
				time.Sleep(time.Second)

				continue
			}

			return nil
		}
	}
}

func (c *cryptoCompareWs) handleWsMessages(ctx context.Context, pipe chan *domain.Data) {
	go func() {
		defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
			if errClose := conn.Close(code, reason); errClose != nil {
				c.l.Println(errClose)
			}
		}(c.conn, websocket.StatusNormalClosure, "")
		var (
			// CryptoCompare automatically send a heartbeat message per socket every 30 seconds,
			// if we miss two heartbeats it means our connection might be stale. So let's start with 2 and
			// add 1 every time the server sends a heartbeat and subtract 1 by ticker every 30 seconds.
			hb   = 2
			tick = time.NewTicker(30 * time.Second)
		)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				if hb <= 0 {
					if err := c.handleWssError(ctx, errHeartbeat); err != nil {
						c.l.Println(err)

						return
					}
				}
				hb--
			default:
				var (
					body []byte
					err  error
				)
				if _, body, err = c.conn.Read(ctx); err != nil {
					if err = c.handleWssError(ctx, err); err != nil {
						c.l.Println(err)

						return
					}

					continue
				}
				data := &cryptoCompareWsData{}
				if err = json.Unmarshal(body, data); err != nil {
					c.l.Println(err)

					continue
				}
				switch data.Type {
				case "999":
					hb++
				case "5":
					pipe <- convertCryptoCompareWsDataToDomain(data)
				}
			}
		}
	}()
}

func convertCryptoCompareWsDataToDomain(d *cryptoCompareWsData) *domain.Data {
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
