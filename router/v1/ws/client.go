package ws

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
	"io"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) pongHandler(string) error {
	return c.conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) readPump() {
	var err error
	defer func() {
		c.hub.unregister <- c
		if err = c.conn.Close(); err != nil {
			handlers.SystemHandler(err)
			return
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err = c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		handlers.SystemHandler(err)
		return
	}
	c.conn.SetPongHandler(c.pongHandler)
	for {
		query := v1.PriceQuery{}
		if err = c.conn.ReadJSON(&query); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				handlers.SystemHandler(err)
				break
			}
			c.errorHandler(errors.New("invalid request: the request should look like {\"fsym\":\"CRYPTO\",\"tsym\":\"COMMON\"}"))
			continue
		}
		c.hub.queryQueue <- &Query{
			send:  c.send,
			query: &query,
		}
	}
}

func (c *Client) writePump() {
	var (
		err    error
		writer io.WriteCloser
		ticker = time.NewTicker(pingPeriod)
	)
	defer func() {
		ticker.Stop()
		if err = c.conn.Close(); err != nil {
			handlers.SystemHandler(err)
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err = c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				handlers.SystemHandler(err)
				return
			}
			if !ok {
				if err = c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					handlers.SystemHandler(err)
					return
				}
				return
			}
			if writer, err = c.conn.NextWriter(websocket.TextMessage); err != nil {
				return
			}
			if _, err = writer.Write(message); err != nil {
				handlers.SystemHandler(err)
				return
			}
			for i := 0; i < len(c.send); i++ {
				if _, err = writer.Write(<-c.send); err != nil {
					handlers.SystemHandler(err)
					return
				}
			}
			if err = writer.Close(); err != nil {
				handlers.SystemHandler(err)
				return
			}
		case <-ticker.C:
			if err = c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				handlers.SystemHandler(err)
				return
			}
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				handlers.SystemHandler(err)
				return
			}
		}
	}
}

// ServeWs - handles websocket requests from the peer.
func ServeWs(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			handlers.SystemHandler(err)
			return
		}
		client := &Client{
			hub:  hub,
			conn: conn,
			send: make(chan []byte, 256),
		}
		client.hub.clients[client] = struct{}{}
		go client.writePump()
		go client.readPump()
	}
}

func (c *Client) errorHandler(err error) {
	var binaryString []byte
	res := handlers.Result{}
	res.UpdateAllFields(http.StatusBadRequest, err.Error(), nil)
	if binaryString, err = json.Marshal(&res); err != nil {
		handlers.SystemHandler(err)
		return
	}
	c.send <- binaryString
}
