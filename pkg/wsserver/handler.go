package ws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/cache"
	v1 "github.com/streamdp/ccd/server/api/v1"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512

	welcomeMessage = "Welcome to CCD WS Server! To get the latest price send request like this: " +
		"{\"type\": \"price\", \n\"pair\": { \"fsym\":\"CRYPTO\",\"tsym\":\"COMMON\" }}"
	closeMessage = "Goodbye!"
)

type client struct {
	handler *wsHandler
	cancel  context.CancelFunc
}

type wsHandler struct {
	l           *log.Logger
	conn        *websocket.Conn
	messagePipe chan []byte

	rc clients.RestClient
	db db.Database

	subscriptions *cache.Cache

	isActive bool
}

var errUnknownMessageType = errors.New("unknown message type")

func (w *wsHandler) handleClientRequests(ctx context.Context) {
	defer close(w.messagePipe)

	defer func() {
		w.isActive = false
	}()

	w.sendMessage(welcomeMessage)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			query := &wsMessage{}
			if err := wsjson.Read(ctx, w.conn, query); err != nil {
				if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				w.l.Println(err)

				return
			}
			query.Pair.toUpper()

			switch query.T {
			case "price":
				data, err := w.getLastPrice(query.Pair)
				if err != nil {
					w.returnClientError(fmt.Errorf("failed to handle get price action: %w", err))
					w.l.Println(err)

					continue
				}
				w.messagePipe <- (&wsMessage{
					T:    "data",
					Data: data,
				}).Marshal()
			case "subscribe":
				w.subscribe(query.Pair)
			case "unsubscribe":
				w.unsubscribe(query.Pair)
			case "close":
				w.sendMessage(closeMessage)
				time.Sleep(3 * time.Second)

				if err := w.Close("closed by client request"); err != nil {
					w.l.Println(err)
				}

				return
			case "ping":
				w.messagePipe <- []byte("pong")
			default:
				w.returnClientError(errUnknownMessageType)
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

func (w *wsHandler) subscribe(p *pair) {
	subscription := p.buildName()
	if w.subscriptions.IsPresent(subscription) {
		w.sendMessage("Already subscribed")

		return
	}

	w.subscriptions.Add(subscription)
	w.sendMessage(fmt.Sprintf("Successfully subscribed on %s/%s pair updates", p.From, p.To))
}

func (w *wsHandler) unsubscribe(p *pair) {
	subscription := p.buildName()
	if !w.subscriptions.IsPresent(subscription) {
		w.sendMessage("Not subscribed")

		return
	}

	w.subscriptions.Remove(subscription)
	w.sendMessage(fmt.Sprintf("Successfully unsubscribed from %s/%s pair updates", p.From, p.To))
}

func (w *wsHandler) sendMessage(message string) {
	w.messagePipe <- (&wsMessage{
		T:       "message",
		Message: message,
	}).Marshal()
}

func (w *wsHandler) returnClientError(err error) {
	w.messagePipe <- (&wsMessage{
		T:       "error",
		Message: err.Error(),
	}).Marshal()
}

func (w *wsHandler) getLastPrice(p *pair) (*domain.Data, error) {
	data, err := v1.LastPrice(w.rc, w.db, p.From, p.To)
	if err != nil {
		return nil, fmt.Errorf("failed to get last price: %w", err)
	}

	return data, nil
}

func (w *wsHandler) Close(reason string) error {
	if err := w.conn.Close(websocket.StatusNormalClosure, reason); err != nil {
		return err
	}

	return nil
}
