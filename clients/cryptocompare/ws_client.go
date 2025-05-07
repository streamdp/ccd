package cryptocompare

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/wsclient"
)

func InitWs(ctx context.Context, pipe chan *domain.Data, l *log.Logger, cfg *config.App) (*wsclient.Ws, error) {
	if cfg.ApiKey == "" {
		return nil, errApiKeyNotDefined
	}

	wssUrl, err := buildURL("wss://streamer.cryptocompare.com/v2", cfg.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %w", err)
	}

	w := wsclient.New(ctx, wssUrl.String(), l)

	w.ChannelNameBuilder = buildChannelName

	w.SubscribeMessageBuilder = func(ch string, id int64) ([]byte, error) {
		return []byte(fmt.Sprintf("{\"action\":\"SubAdd\",\"subs\":[\"%s\"]}", ch)), nil
	}

	w.UnsubscribeMessageBuilder = func(ch string, id int64) ([]byte, error) {
		return []byte(fmt.Sprintf("{\"action\":\"SubRemove\",\"subs\":[\"%s\"]}", ch)), nil
	}

	w.MessageHandler = func(ctx context.Context) {
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
					if err := w.HandleWsError(ctx, errHeartbeat); err != nil {
						l.Println(err)

						return
					}
				}
				hb.decrease()
			default:
				body, err := w.Read(ctx)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						l.Println(err)
					}

					return
				}

				data := &wsData{}
				if err := json.Unmarshal(body, data); err != nil {
					l.Println(err)

					continue
				}
				switch data.Type {
				case "999":
					hb.reset()
				case "3", "16", "17", "18", "429", "500":
					if msg := handleServerResponse(data); msg != "" {
						l.Println(msg)
					}
				case "5":
					pipe <- convertWsDataToDomain(data)
				}
			}
		}
	}

	return w, nil
}

func buildURL(wssUrl string, apiKey string) (*url.URL, error) {
	u, err := url.Parse(wssUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	q := u.Query()
	q.Set("api_key", apiKey)
	u.RawQuery = q.Encode()

	return u, nil
}

func buildChannelName(from, to string) string {
	return fmt.Sprintf("5~CCCAGG~%s~%s", strings.ToUpper(from), strings.ToUpper(to))
}

func handleServerResponse(data *wsData) string {
	switch data.Message {
	case "SUBSCRIBECOMPLETE":
		return "cccagg channel: successfully subscribed on the " + data.Sub
	case "UNSUBSCRIBECOMPLETE":
		return "cccagg channel: successfully unsubscribed from the " + data.Sub
	case "UNSUBSCRIBEALLCOMPLETE":
		return "cccagg channel: " + data.Info
	case "LOADCOMPLETE":
		return "cccagg channel: " + data.Info
	case "SUBSCRIPTION_ALREADY_ACTIVE":
		return fmt.Sprintf("cccagg channel: %s subscription already active", data.Parameter)
	case "TOO_MANY_SOCKETS_MAX_1_PER_CLIENT":
		return "cccagg channel: " + data.Info
	}

	return ""
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
