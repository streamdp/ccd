package cryptocompare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/handlers"
)

const wssUrl = "wss://streamer.cryptocompare.com/v2"

type cryptoCompareWsData struct {
	Type                string  `json:"TYPE"`
	FromSymbol          string  `json:"FROMSYMBOL"`
	ToSymbol            string  `json:"TOSYMBOL"`
	Price               float64 `json:"PRICE"`
	LastUpdate          int64   `json:"LASTUPDATE"`
	Volume24Hour        float64 `json:"VOLUME24HOUR"`
	Volume24HourTo      float64 `json:"VOLUME24HOURTO"`
	Open24Hour          float64 `json:"OPEN24HOUR"`
	High24Hour          float64 `json:"HIGH24HOUR"`
	Low24Hour           float64 `json:"LOW24HOUR"`
	CurrentSupply       float64 `json:"CURRENTSUPPLY"`
	CurrentSupplyMktCap float64 `json:"CURRENTSUPPLYMKTCAP"`
}

type cryptoCompareWs struct {
	ctx        context.Context
	conn       *websocket.Conn
	subscribes clients.Subscribes
	subMu      sync.RWMutex
}

func InitWs(pipe chan *clients.Data) (clients.WsClient, error) {
	h := &cryptoCompareWs{
		ctx:        context.Background(),
		subscribes: clients.Subscribes{},
	}
	if err := h.reconnect(); err != nil {
		return nil, err
	}
	h.handleWsMessages(pipe)
	return h, nil
}

func (c *cryptoCompareWs) reconnect() (err error) {
	if c.conn != nil {
		if err := c.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			handlers.SystemHandler(err)
		}
	}
	var u *url.URL
	if u, err = c.buildURL(); err != nil {
		return
	}
	c.conn, _, err = websocket.Dial(c.ctx, u.String(), nil)
	return
}

func (c *cryptoCompareWs) buildURL() (u *url.URL, err error) {
	var apiKey string
	if apiKey, err = getApiKey(); err != nil {
		return
	}
	if u, err = url.Parse(wssUrl); err != nil {
		return
	}
	query := u.Query()
	query.Set("api_key", apiKey)
	u.RawQuery = query.Encode()
	return
}

func (c *cryptoCompareWs) resubscribe() (err error) {
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	for k := range c.subscribes {
		if err = c.sendSubscribeMsg(k); err != nil {
			return
		}
	}
	return
}

func (c *cryptoCompareWs) handleWssError(err error) error {
	handlers.SystemHandler(err)
	for {
		select {
		case <-time.After(time.Minute):
			return errors.New("reconnect failed")
		default:
			if err = c.reconnect(); err != nil {
				time.Sleep(time.Second)
				continue
			}
			if err = c.resubscribe(); err != nil {
				time.Sleep(time.Second)
				continue
			}
			return nil
		}
	}
}

func (c *cryptoCompareWs) handleWsMessages(pipe chan *clients.Data) {
	go func() {
		defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
			if err := conn.Close(code, reason); err != nil {
				handlers.SystemHandler(err)
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
			case <-c.ctx.Done():
				return
			case <-tick.C:
				if hb <= 0 {
					if err := c.handleWssError(errors.New("heartbeat loss")); err != nil {
						handlers.SystemHandler(err)
						return
					}
				}
				hb--
			default:
				var (
					body []byte
					err  error
				)
				if _, body, err = c.conn.Read(c.ctx); err != nil {
					if err = c.handleWssError(err); err != nil {
						handlers.SystemHandler(err)
						return
					}
					continue
				}
				data := &cryptoCompareWsData{}
				if err = json.Unmarshal(body, data); err != nil {
					handlers.SystemHandler(err)
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

func buildChannelName(from, to string) string {
	return fmt.Sprintf("5~CCCAGG~%s~%s", strings.ToUpper(from), strings.ToUpper(to))
}

func (c *cryptoCompareWs) Unsubscribe(from, to string) (err error) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if _, ok := c.subscribes[ch]; ok {
		if err = c.sendUnsubscribeMsg(ch); err != nil {
			return
		}
		delete(c.subscribes, ch)
	}
	return
}

func (c *cryptoCompareWs) sendUnsubscribeMsg(ch string) error {
	return c.conn.Write(c.ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubRemove\",\"subs\":[\"%s\"]}", ch)),
	)
}

func (c *cryptoCompareWs) Subscribe(from, to string) (err error) {
	c.subMu.Lock()
	defer c.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if err = c.sendSubscribeMsg(ch); err != nil {
		return
	}
	c.subscribes[ch] = clients.NewSubscribe(from, to, 0)
	return
}

func (c *cryptoCompareWs) sendSubscribeMsg(ch string) error {
	return c.conn.Write(c.ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"action\":\"SubAdd\",\"subs\":[\"%s\"]}", ch)),
	)
}

func (c *cryptoCompareWs) ListSubscribes() clients.Subscribes {
	s := make(clients.Subscribes, len(c.subscribes))
	c.subMu.RLock()
	defer c.subMu.RUnlock()
	for k, v := range c.subscribes {
		s[k] = v
	}
	return s
}

func convertCryptoCompareWsDataToDomain(d *cryptoCompareWsData) *clients.Data {
	if d == nil {
		return nil
	}
	return &clients.Data{
		From: d.FromSymbol,
		To:   d.ToSymbol,
		Raw: &clients.Response{
			Open24Hour:     d.Open24Hour,
			Volume24Hour:   d.Volume24Hour,
			Volume24Hourto: d.Volume24HourTo,
			Low24Hour:      d.Low24Hour,
			High24Hour:     d.High24Hour,
			Price:          d.Price,
			Supply:         d.CurrentSupply,
			MktCap:         d.CurrentSupplyMktCap,
			LastUpdate:     d.LastUpdate,
		},
		Display: &clients.Display{
			Open24Hour:     strconv.FormatFloat(d.Open24Hour, 'f', -1, 64),
			Volume24Hour:   strconv.FormatFloat(d.Volume24Hour, 'f', -1, 64),
			Volume24Hourto: strconv.FormatFloat(d.Volume24HourTo, 'f', -1, 64),
			High24Hour:     strconv.FormatFloat(d.High24Hour, 'f', -1, 64),
			Price:          strconv.FormatFloat(d.Price, 'f', -1, 64),
			FromSymbol:     strings.ToUpper(d.FromSymbol),
			ToSymbol:       strings.ToUpper(d.ToSymbol),
			LastUpdate:     strconv.FormatInt(d.LastUpdate, 10),
			Supply:         strconv.FormatFloat(d.CurrentSupply, 'f', -1, 64),
			MktCap:         strconv.FormatFloat(d.CurrentSupplyMktCap, 'f', -1, 64),
		},
	}
}
