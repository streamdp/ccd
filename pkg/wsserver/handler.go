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
	welcomeMessage = "Welcome to the CCD Server! To get the latest price send request like this: " +
		"{\"type\": \"price\", \n\"pair\": { \"fsym\":\"CRYPTO\",\"tsym\":\"COMMON\" }}"
	closeMessage = "Connection will be closed at the client's request."
)

const (
	readWait       = time.Minute
	writeWait      = 10 * time.Second
	maxMessageSize = 512

	messageTypeMessage = "message"
	messageTypeError   = "error"
)

type handler struct {
	l           *log.Logger
	conn        *websocket.Conn
	messagePipe chan []byte

	rc clients.RestClient
	db db.Database

	subscriptions *cache.Cache

	isActive bool
}

func (h *handler) handleClientRequests(ctx context.Context) {
	defer close(h.messagePipe)

	h.sendMessage(messageTypeMessage, welcomeMessage)

	h.isActive = true
	defer func() {
		h.isActive = false
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg := &wsMessage{}

			ctxRead, cancel := ctx, context.CancelFunc(func() {})
			if h.subscriptions.Len() == 0 {
				ctxRead, cancel = context.WithTimeout(ctx, readWait)
			}
			if err := wsjson.Read(ctxRead, h.conn, msg); err != nil {
				cancel()
				if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				h.l.Println(err)

				return
			}
			cancel()

			if msg.Pair != nil {
				msg.Pair.toUpper()
			}

			switch msg.T {
			case "price":
				data, err := h.getLastPrice(msg.Pair)
				if err != nil {
					h.sendMessage(messageTypeError, "failed to handle get price action: "+err.Error())
					h.l.Println(err)

					continue
				}
				h.messagePipe <- (&wsMessage{
					T:         "data",
					Data:      data,
					Timestamp: time.Now().UTC().UnixMilli(),
				}).Bytes()
			case "subscribe":
				h.subscribe(msg.Pair)
			case "unsubscribe":
				h.unsubscribe(msg.Pair)
			case "close":
				h.sendMessage(messageTypeMessage, closeMessage)
				time.Sleep(3 * time.Second)

				if err := h.close("closed at the client's request"); err != nil {
					h.l.Println(err)
				}

				return
			case "ping":
				h.messagePipe <- (&wsMessage{
					T:         "pong",
					Timestamp: msg.Timestamp,
				}).Bytes()
			default:
				h.sendMessage(messageTypeError, "unknown message type")
			}
		}
	}
}

func (h *handler) handleMessagePipe(ctx context.Context) {
	for message := range h.messagePipe {
		ctx, cancel := context.WithTimeout(ctx, writeWait)
		if err := h.conn.Write(ctx, websocket.MessageText, message); err != nil {
			h.l.Println(err)
			cancel()

			return
		}
		cancel()
	}
}

func (h *handler) subscribe(p *pair) {
	subscription := p.buildName()
	if h.subscriptions.IsPresent(subscription) {
		h.sendMessage(messageTypeMessage, "Already subscribed")

		return
	}

	h.subscriptions.Add(subscription)
	h.sendMessage(
		messageTypeMessage,
		fmt.Sprintf("Successfully subscribed on %s/%s pair updates", p.From, p.To),
	)
}

func (h *handler) unsubscribe(p *pair) {
	subscription := p.buildName()
	if !h.subscriptions.IsPresent(subscription) {
		h.sendMessage(messageTypeMessage, "Not subscribed")

		return
	}

	h.subscriptions.Remove(subscription)
	h.sendMessage(
		messageTypeMessage,
		fmt.Sprintf("Successfully unsubscribed from %s/%s pair updates", p.From, p.To),
	)
}

func (h *handler) sendMessage(messageType, message string) {
	h.messagePipe <- (&wsMessage{
		T:         messageType,
		Message:   message,
		Timestamp: time.Now().UTC().UnixMilli(),
	}).Bytes()
}

func (h *handler) getLastPrice(p *pair) (*domain.Data, error) {
	data, err := v1.LastPrice(h.rc, h.db, p.From, p.To)
	if err != nil {
		return nil, fmt.Errorf("failed to get last price: %w", err)
	}

	return data, nil
}

func (h *handler) close(reason string) error {
	if err := h.conn.Close(websocket.StatusNormalClosure, reason); err != nil {
		return fmt.Errorf("failed to close ws connection: %w", err)
	}

	return nil
}
