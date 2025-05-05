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
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
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
}

var errInvalidRequest = errors.New(
	`invalid request: request should look like {"fsym":"CRYPTO","tsym":"COMMON"}`,
)

// HandleWs - handles websocket requests from the peer.
func HandleWs(ctx context.Context, r clients.RestClient, l *log.Logger, db db.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			l.Println(err)

			return
		}

		h := &wsHandler{
			l:           l,
			conn:        conn,
			messagePipe: make(chan []byte, 256),
			rc:          r,
			db:          db,
		}
		h.conn.SetReadLimit(maxMessageSize)

		go h.handleMessagePipe(ctx)
		go h.handleClientRequests(ctx)
	}
}

func (w *wsHandler) handleClientRequests(ctx context.Context) {
	defer func() {
		close(w.messagePipe)
	}()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var (
				data  []byte
				err   error
				query = v1.PriceQuery{}
			)

			if _, data, err = w.conn.Read(ctx); err != nil {
				if errors.As(err, &websocket.CloseError{}) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				w.l.Println(err)

				return
			}

			if err = json.Unmarshal(data, &query); err != nil {
				w.returnAnErrorToTheClient(errInvalidRequest)

				continue
			}
			query.ToUpper()

			if data, err = w.getLastPrice(query.From, query.To); err != nil {
				w.l.Println(err)

				continue
			}

			w.messagePipe <- data
		}
	}
}

func (w *wsHandler) getLastPrice(from, to string) ([]byte, error) {
	data, err := v1.LastPrice(w.rc, w.db, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get last price: %w", err)
	}

	result, errMarshal := json.Marshal(&data)
	if errMarshal != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", errMarshal)
	}

	return result, nil
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

func (w *wsHandler) returnAnErrorToTheClient(err error) {
	var binaryString []byte
	r := domain.NewResult(http.StatusBadRequest, err.Error(), nil)
	if binaryString, err = json.Marshal(&r); err != nil {
		w.l.Println(err)

		return
	}
	w.messagePipe <- binaryString
}
