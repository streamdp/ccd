package kraken

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/wsclient"
)

const defaultPingInterval = 5 * time.Second

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger) *wsclient.Ws {
	w := wsclient.New(ctx, "wss://ws.kraken.com/v2", l)

	w.ChannelNameBuilder = buildChannelName

	w.SubscribeMessageBuilder = func(ch string, _ int64) ([]byte, error) {
		return buildWsMessage("subscribe", []string{ch})
	}

	w.UnsubscribeMessageBuilder = func(ch string, _ int64) ([]byte, error) {
		return buildWsMessage("unsubscribe", []string{ch})
	}

	w.PingMessageBuilder = func(ch string, id int64) ([]byte, error) {
		msg, _ := json.Marshal(wsMessage{
			Method: "ping",
			ReqId:  id,
		})

		return msg, nil
	}

	w.MessageHandler = func(ctx context.Context) {
		defer w.WsDown(true)

		t := time.NewTimer(defaultPingInterval)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := w.Ping(ctx, "", time.Now().UTC().UnixMilli()); err != nil {
					l.Println(err)
				}
				t.Reset(defaultPingInterval)
			default:
				body, err := w.Read(ctx)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						l.Println(err)
					}
					if errors.Is(err, wsclient.ErrClientReconnected) {
						continue
					}

					return
				}

				if bytes.Contains(body, []byte("method")) {
					if msg := handleServerResponse(body); msg != "" {
						l.Println(msg)
					}

					continue
				}

				if bytes.Contains(body, []byte("heartbeat")) {
					continue
				}

				if err = handleWsUpdate(w, pipe, body); err != nil {
					l.Println(err)
				}
			}
		}
	}

	return w
}

func buildChannelName(from, to string) string {
	return fmt.Sprintf("%s/%s", strings.ToUpper(from), strings.ToUpper(to))
}

func buildWsMessage(method string, channels []string) ([]byte, error) {
	msg, err := json.Marshal(wsMessage{
		Method: method,
		Params: &wsMessageParams{
			Channel: "ticker",
			Symbol:  channels,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	return msg, nil
}

func handleServerResponse(body []byte) string {
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
	case "pong":
		return ""
	}

	return ""
}

func handleWsUpdate(w *wsclient.Ws, pipe chan *domain.Data, body []byte) error {
	data := &wsData{}
	if err := json.Unmarshal(body, data); err != nil {
		return fmt.Errorf("failed to unmarshal ws update message: %w", err)
	}

	if data.Channel != "ticker" || len(data.Data) == 0 {
		return nil
	}

	for _, tick := range data.Data {
		from, to := w.PairFromChannelName(tick.Symbol)
		if from == "" || to == "" {
			continue
		}

		pipe <- convertWsDataToDomain(from, to, &tick, time.Now().UTC().UnixMilli())
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
