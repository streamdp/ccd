package huobi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
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
	subMu      sync.Mutex
}

func InitWs(pipe chan *clients.Data) clients.WssClient {
	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, wssUrl, nil)
	if err != nil {
		handlers.SystemHandler(err)
	}
	h := &huobiWs{
		ctx:        ctx,
		conn:       conn,
		subscribes: map[string]*channel{},
	}
	h.handleWsMessages(pipe)
	return h
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
					handlers.SystemHandler(err)
					return
				}
				if body, err = gzipDecompress(r); err != nil {
					handlers.SystemHandler(err)
					continue
				}
				if bytes.Contains(body, []byte("ping")) {
					body = bytes.Replace(body, []byte("ping"), []byte("pong"), -1)
					if err = h.conn.Write(h.ctx, websocket.MessageText, body); err != nil {
						handlers.SystemHandler(err)
						return
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
				from, to := h.getPair(data.Ch)
				if from != "" && to != "" {
					pipe <- convertHuobiWssDataToDomain(from, to, data)
				}
			}
		}
	}()
}

func (h *huobiWs) getPair(ch string) (from, to string) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	s := strings.Split(ch, ".")
	if len(s) != 3 {
		return
	}
	if c, ok := h.subscribes[s[1]]; ok {
		return c.from, c.to
	}
	return
}

func buildSymbol(from, to string) string {
	if strings.ToLower(to) == "usd" {
		to = "usdt"
	}
	return strings.ToLower(from + to)
}

func (h *huobiWs) UnSubscribe(from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var symbol = buildSymbol(from, to)
	if c, ok := h.subscribes[symbol]; ok {
		if err = h.conn.Write(h.ctx, websocket.MessageText, []byte(
			fmt.Sprintf("{\"unsub\": \"market.%s.ticker\", \"id\":\"%d\"}", symbol, c.id)),
		); err != nil {
			return
		}
		delete(h.subscribes, symbol)
	}
	return
}

func (h *huobiWs) Subscribe(from, to string) (err error) {
	h.subMu.Lock()
	defer h.subMu.Unlock()
	var (
		id     = time.Now().UnixMilli()
		symbol = buildSymbol(from, to)
	)
	if err = h.conn.Write(h.ctx, websocket.MessageText, []byte(
		fmt.Sprintf("{\"sub\": \"market.%s.ticker\", \"id\":\"%d\"}", symbol, id)),
	); err != nil {
		return
	}
	h.subscribes[symbol] = &channel{
		from: strings.ToUpper(from),
		to:   strings.ToUpper(to),
		id:   id,
	}
	return
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
