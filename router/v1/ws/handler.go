package ws

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

type wsHandler struct {
	ctx         context.Context
	cancel      context.CancelFunc
	conn        *websocket.Conn
	messagePipe chan []byte

	rc clients.RestClient
	db db.DbReadWrite
}

// HandleWs - handles websocket requests from the peer.
func HandleWs(rc clients.RestClient, db db.DbReadWrite) gin.HandlerFunc {
	return func(c *gin.Context) {
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		ctx, cancel := context.WithCancel(context.Background())
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			cancel()
			handlers.SystemHandler(err)
			return
		}
		h := &wsHandler{
			ctx:         ctx,
			cancel:      cancel,
			conn:        conn,
			messagePipe: make(chan []byte, 256),
			rc:          rc,
			db:          db,
		}
		h.conn.SetReadLimit(maxMessageSize)
		h.conn.SetPingHandler(h.pingHandler)
		h.conn.SetPongHandler(h.pongHandler)
		go h.handleMessagePipe()
		go h.handleClientRequests()
	}
}

func (w *wsHandler) pingHandler(string) error {
	if err := w.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return w.conn.WriteMessage(websocket.PongMessage, nil)
}

func (w *wsHandler) pongHandler(string) error {
	if err := w.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return w.conn.WriteMessage(websocket.PingMessage, nil)
}

func (w *wsHandler) handleClientRequests() {
	defer func() {
		w.cancel()
		close(w.messagePipe)
	}()
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			var (
				data  []byte
				err   error
				query = v1.PriceQuery{}
			)
			if err = w.conn.ReadJSON(&query); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					handlers.SystemHandler(err)
					return
				}
				w.returnAnErrorToTheClient(errors.New(
					"invalid request: the request should look like {\"fsym\":\"CRYPTO\",\"tsym\":\"COMMON\"}",
				))
				continue
			}
			if data, err = w.getLastPrice(&query); err != nil {
				handlers.SystemHandler(err)
				return
			}
			w.messagePipe <- data
		}
	}
}

func (w *wsHandler) getLastPrice(query *v1.PriceQuery) (result []byte, err error) {
	var data *clients.Data
	if data, err = v1.GetLastPrice(w.rc, w.db, query); err != nil {
		return
	}
	if result, err = json.Marshal(&data); err != nil {
		return
	}
	return
}

func (w *wsHandler) handleMessagePipe() {
	defer w.cancel()
	for message := range w.messagePipe {
		var (
			writer io.WriteCloser
			err    error
		)
		if err = w.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			handlers.SystemHandler(err)
			return
		}
		if writer, err = w.conn.NextWriter(websocket.TextMessage); err != nil {
			handlers.SystemHandler(err)
			return
		}
		if _, err = writer.Write(message); err != nil {
			handlers.SystemHandler(err)
			return
		}
		if err = writer.Close(); err != nil {
			handlers.SystemHandler(err)
			return
		}
	}
	if err := w.conn.Close(); err != nil {
		handlers.SystemHandler(err)
		return
	}
}

func (w *wsHandler) returnAnErrorToTheClient(err error) {
	var binaryString []byte
	res := handlers.Result{}
	res.UpdateAllFields(http.StatusBadRequest, err.Error(), nil)
	if binaryString, err = json.Marshal(&res); err != nil {
		handlers.SystemHandler(err)
		return
	}
	w.messagePipe <- binaryString
}
