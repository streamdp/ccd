package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/cache"
	v1 "github.com/streamdp/ccd/server/api/v1"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

type wsHandler struct {
	l           *log.Logger
	conn        *websocket.Conn
	messagePipe chan []byte

	rc clients.RestClient
	db db.Database

	subscriptions *cache.Cache

	isActive bool
}

var errInvalidRequest = errors.New(
	`invalid request: request should look like {"fsym":"CRYPTO","tsym":"COMMON"}`,
)

func (w *wsHandler) handleClientRequests(ctx context.Context) {
	defer close(w.messagePipe)

	t := time.NewTimer(time.Minute)
	defer t.Stop()

	defer func() {
		w.isActive = false
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if w.subscriptions.Len() != 0 {
				continue
			}
			w.isActive = false

			return
		default:
			var (
				data  []byte
				err   error
				query = wsMessage{}
			)

			if _, data, err = w.conn.Read(ctx); err != nil {
				if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				w.l.Println(err)

				return
			}

			b, err := json.Marshal(&wsMessage{
				T: "price",
				Pair: pair{
					From: "BTC",
					To:   "USDT",
				},
			})
			w.l.Println(string(b))

			if err = json.Unmarshal(data, &query); err != nil {
				w.returnAnErrorToTheClient(errInvalidRequest)

				continue
			}
			query.Pair.toUpper()

			switch query.T {
			case "price":
				if data, err = w.getLastPrice(&query.Pair); err != nil {
					w.l.Println(err)

					continue
				}
				w.messagePipe <- data
			case "subscribe":
				subscriptionName := query.Pair.buildName()
				if w.subscriptions.IsPresent(subscriptionName) {
					w.messagePipe <- []byte("Already subscribed")

					continue
				}
				w.subscriptions.Add(subscriptionName)
				w.messagePipe <- []byte(fmt.Sprintf(
					`Successfully subscribed on %s/%s pair updates`, query.Pair.From, query.Pair.To,
				))
			case "unsubscribe":
				subscriptionName := query.Pair.buildName()
				if !w.subscriptions.IsPresent(subscriptionName) {
					w.messagePipe <- []byte("Not subscribed")

					continue
				}
				w.subscriptions.Remove(subscriptionName)
				w.messagePipe <- []byte(fmt.Sprintf(
					`Successfully unsubscribed from %s/%s pair updates`, query.Pair.From, query.Pair.To,
				))
			case "close":
				w.isActive = false
				w.l.Println("Client send close message, reason: " + query.Reason)
				w.sendCloseMessage()
				time.Sleep(time.Second)

				return
			case "ping":
				t.Reset(time.Minute)
				w.messagePipe <- []byte("pong")
			default:
				w.messagePipe <- []byte("Unknown message type: " + query.T)
			}
		}
	}
}

func (w *wsHandler) handleMessagePipe(ctx context.Context) {
	for message := range w.messagePipe {
		ctx, cancel := context.WithTimeout(ctx, writeWait)
		if err := w.conn.Write(ctx, websocket.MessageText, message); err != nil {
			w.l.Println(err)
			cancel()

			return
		}
		cancel()
	}
}

func (w *wsHandler) sendWelcomeMessage() {
	w.messagePipe <- []byte(`Welcome to CCD WS Server!`)
	w.messagePipe <- []byte(`To get the latest price send request like this:`)
	w.messagePipe <- []byte(`{"type": "price", pair:{"fsym":"CRYPTO","tsym":"COMMON"}}`)
}

func (w *wsHandler) sendCloseMessage() {
	w.messagePipe <- []byte(`Goodbye!`)
}

func (w *wsHandler) returnAnErrorToTheClient(err error) {
	var binaryString []byte
	r := domain.NewResult(http.StatusBadRequest, err.Error(), nil)
	if binaryString, err = json.Marshal(&r); err != nil {
		w.l.Println(err)

		return
	}
	w.messagePipe <- binaryString
}

func (w *wsHandler) getLastPrice(p *pair) ([]byte, error) {
	data, err := v1.LastPrice(w.rc, w.db, p.From, p.To)
	if err != nil {
		return nil, fmt.Errorf("failed to get last price: %w", err)
	}

	result, errMarshal := json.Marshal(&data)
	if errMarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", errMarshal)
	}

	return result, nil
}
