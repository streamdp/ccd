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

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/wsclient"
)

func InitWs(
	ctx context.Context,
	pipe chan *domain.Data,
	sessionRepo clients.SessionRepo,
	l *log.Logger,
	cfg *config.Http,
) *wsclient.Ws {
	w := wsclient.New(ctx, "wss://api.huobi.pro/ws", sessionRepo, l, cfg)

	w.ChannelNameBuilder = buildChannelName

	w.SubscribeMessageBuilder = func(ch string, id int64) ([]byte, error) {
		return []byte(fmt.Sprintf("{\"sub\": \"%s\", \"id\":\"%d\"}", ch, id)), nil
	}

	w.UnsubscribeMessageBuilder = func(ch string, id int64) ([]byte, error) {
		return []byte(fmt.Sprintf("{\"unsub\": \"%s\", \"id\":\"%d\"}", ch, id)), nil
	}

	w.PongMessageBuilder = func(ch string, id int64) ([]byte, error) {
		return []byte(fmt.Sprintf("{\"pong\":%d}", id)), nil
	}

	w.MessageHandler = func(ctx context.Context) {
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
				if r, err = w.Reader(ctx); err != nil {
					if !errors.Is(err, context.Canceled) {
						l.Println(err)
					}
					if errors.Is(err, context.Canceled) || errors.Is(err, wsclient.ErrClientReconnected) {
						continue
					}
					w.WsDown()

					return
				}

				if body, err = gzipDecompress(r); err != nil {
					l.Println(err)

					continue
				}
				if bytes.Contains(body, []byte("ping")) {
					if err = w.Pong(ctx, "", (&wsPing{}).Unmarshal(body).Ts); err != nil {
						l.Println(err)
					}

					continue
				}
				if bytes.Contains(body, []byte("subbed")) {
					if msg := handleServerResponse(body); msg != "" {
						l.Println(msg)
					}

					continue
				}
				if err = handleWsUpdate(w, body, pipe); err != nil {
					l.Println(err)
				}
			}
		}
	}

	return w
}

func buildChannelName(from, to string) string {
	if strings.ToLower(to) == "usd" {
		to = "usdt"
	}

	return fmt.Sprintf("market.%s.ticker", strings.ToLower(from+to))
}

func handleServerResponse(body []byte) string {
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

func handleWsUpdate(w *wsclient.Ws, body []byte, pipe chan *domain.Data) error {
	data := &wsData{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("failed to unmarshal ws update message: %w", err)
	}

	if data.Ch == "" {
		return nil
	}

	from, to := w.PairFromChannelName(data.Ch)
	if from == "" || to == "" {
		return nil
	}

	pipe <- convertWsDataToDomain(from, to, data)

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
