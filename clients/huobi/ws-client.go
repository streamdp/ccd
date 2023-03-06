package huobi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"nhooyr.io/websocket"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/handlers"
)

const wssUrl = "wss://api.huobi.pro/ws"

type huobiWssData struct {
	Ch   string `json:"ch"`
	Ts   int64  `json:"ts"`
	Tick struct {
		Open      float64 `json:"open"`
		High      float64 `json:"high"`
		Low       float64 `json:"low"`
		Close     float64 `json:"close"`
		Amount    float64 `json:"amount"`
		Vol       float64 `json:"vol"`
		Count     int     `json:"count"`
		Bid       float64 `json:"bid"`
		BidSize   float64 `json:"bidSize"`
		Ask       float64 `json:"ask"`
		AskSize   float64 `json:"askSize"`
		LastPrice float64 `json:"lastPrice"`
		LastSize  float64 `json:"lastSize"`
	} `json:"tick"`
}

type channel struct {
	from, to string
	id       int64
}

type huobiWs struct {
	ctx        context.Context
	conn       *websocket.Conn
	subscribes map[string]*channel
	subMu      sync.RWMutex
}

func InitWs(pipe chan *clients.Data) clients.WssClient {
	h := &huobiWs{
		ctx:        context.Background(),
		subscribes: map[string]*channel{},
	}
	if err := h.reconnect(); err != nil {
		handlers.SystemHandler(err)
	}
	h.handleWsMessages(pipe)
	return h
}

func (h *huobiWs) reconnect() (err error) {
	if h.conn != nil {
		if err := h.conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			handlers.SystemHandler(err)
		}
	}
	h.conn, _, err = websocket.Dial(h.ctx, wssUrl, nil)
	if err != nil {
		return
	}
	return
}

func (h *huobiWs) resubscribe() (err error) {
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	for k, v := range h.subscribes {
		if err = h.sendSubscribeMsg(k, v.id); err != nil {
			return
		}
	}
	return
}

func (h *huobiWs) handleWssError(err error) error {
	handlers.SystemHandler(err)
	for {
		select {
		case <-time.After(time.Minute):
			return errors.New("reconnect failed")
		default:
			if err = h.reconnect(); err != nil {
				time.Sleep(time.Second)
				continue
			}
			if err = h.resubscribe(); err != nil {
				time.Sleep(time.Second)
				continue
			}
			return nil
		}
	}
}

func (h *huobiWs) handleWsMessages(pipe chan *clients.Data) {
	go func() {
		defer func(conn *websocket.Conn, code websocket.StatusCode, reason string) {
			if err := conn.Close(code, reason); err != nil {
				handlers.SystemHandler(err)
			}
		}(h.conn, websocket.StatusNormalClosure, "")
		for {
			select {
			case <-h.ctx.Done():
				return
			default:
				var (
					r    io.Reader
					body []byte
					err  error
				)
				if _, r, err = h.conn.Reader(h.ctx); err != nil {
					if err = h.handleWssError(err); err != nil {
						handlers.SystemHandler(err)
						return
					}
					continue
				}
				if body, err = gzipDecompress(r); err != nil {
					handlers.SystemHandler(err)
					continue
				}
				if bytes.Contains(body, []byte("ping")) {
					if err = h.pingHandler(body); err != nil {
						if err = h.handleWssError(err); err != nil {
							handlers.SystemHandler(err)
							return
						}
					}
					continue
				}
				data := &huobiWssData{}
				if err = json.Unmarshal(body, data); err != nil {
					handlers.SystemHandler(err)
					continue
				}
				if data.Ch == "" {
					continue
				}
				from, to := h.pairFromChannelName(data.Ch)
				if from != "" && to != "" {
					pipe <- convertHuobiWssDataToDomain(from, to, data)
				}
			}
		}
	}()
}

func (h *huobiWs) pingHandler(m []byte) (err error) {
	m = bytes.Replace(m, []byte("ping"), []byte("pong"), -1)
	return h.conn.Write(h.ctx, websocket.MessageText, m)
}

func (h *huobiWs) pairFromChannelName(ch string) (from, to string) {
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	if c, ok := h.subscribes[ch]; ok {
		return c.from, c.to
	}
	return
}

func buildChannelName(from, to string) string {
	if strings.ToLower(to) == "usd" {
		to = "usdt"
	}
	return fmt.Sprintf("market.%s.ticker", strings.ToLower(from+to))
}

func (h *huobiWs) Unsubscribe(from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var ch = buildChannelName(from, to)
	if c, ok := h.subscribes[ch]; ok {
		if err = h.sendUnsubscribeMsg(ch, c.id); err != nil {
			return
		}
		delete(h.subscribes, ch)
	}
	return
}

func (h *huobiWs) sendUnsubscribeMsg(ch string, id int64) error {
	return h.conn.Write(h.ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"unsub\": \"%s\", \"id\":\"%d\"}", ch, id)),
	)
}

func (h *huobiWs) Subscribe(from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var (
		id = time.Now().UnixMilli()
		ch = buildChannelName(from, to)
	)
	if err = h.sendSubscribeMsg(ch, id); err != nil {
		return
	}
	h.subscribes[ch] = &channel{
		from: strings.ToUpper(from),
		to:   strings.ToUpper(to),
		id:   id,
	}
	return
}

func (h *huobiWs) sendSubscribeMsg(ch string, id int64) error {
	return h.conn.Write(h.ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"sub\": \"%s\", \"id\":\"%d\"}", ch, id)),
	)
}

func (h *huobiWs) ListSubscribes() clients.Subscribes {
	s := make(clients.Subscribes, len(h.subscribes))
	h.subMu.RLock()
	defer h.subMu.RUnlock()
	for k, v := range h.subscribes {
		s[k] = &clients.Subscribe{
			From: v.from,
			To:   v.to,
		}
	}
	return s
}

func gzipDecompress(r io.Reader) ([]byte, error) {
	r, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

func convertHuobiWssDataToDomain(from, to string, d *huobiWssData) *clients.Data {
	if d == nil {
		return nil
	}
	return &clients.Data{
		From: from,
		To:   to,
		Raw: &clients.Response{
			Open24Hour:     d.Tick.Open,
			Volume24Hour:   d.Tick.Amount,
			Volume24Hourto: d.Tick.Vol,
			Low24Hour:      d.Tick.Low,
			High24Hour:     d.Tick.High,
			Price:          d.Tick.Bid,
			Supply:         float64(d.Tick.Count),
			Lastupdate:     d.Ts,
		},
		Display: &clients.Display{
			Open24Hour:     strconv.FormatFloat(d.Tick.Open, 'f', -1, 64),
			Volume24Hour:   strconv.FormatFloat(d.Tick.Amount, 'f', -1, 64),
			Volume24Hourto: strconv.FormatFloat(d.Tick.Vol, 'f', -1, 64),
			High24Hour:     strconv.FormatFloat(d.Tick.High, 'f', -1, 64),
			Price:          strconv.FormatFloat(d.Tick.Bid, 'f', -1, 64),
			FromSymbol:     strings.ToUpper(from),
			ToSymbol:       strings.ToUpper(to),
			Lastupdate:     strconv.FormatInt(d.Ts, 10),
			Supply:         strconv.Itoa(d.Tick.Count),
		},
	}
}
