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
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub  *hub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) pongHandler(string) (err error) {
	if err = c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}
	return
}

func (c *Client) readPump() {
	var err error
	defer func() {
		c.hub.unregister <- c
		if err = c.conn.Close(); err != nil {
			return
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err = c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(c.pongHandler)
	for {
		query := v1.PriceQuery{}
		if err = c.conn.ReadJSON(&query); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				handlers.SystemHandler(err)
			}
			break
		}
		if query != (v1.PriceQuery{}) {
			c.hub.queryQueue <- &Query{
				sender: c,
				query:  &query,
			}
		} else {
			c.errorHandler(errors.New("invalid request: the request should look like {'fsym':'CRYPTO','tsym':'COMMON'}"))
		}
	}
}

func (c *Client) writePump() {
	var err error
	var writer io.WriteCloser
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err = c.conn.Close(); err != nil {
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			if err = c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if !ok {
				if err = c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					return
				}
				return
			}
			if writer, err = c.conn.NextWriter(websocket.TextMessage); err != nil {
				return
			}
			if _, err = writer.Write(message); err != nil {
				return
			}
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err = writer.Write(<-c.send); err != nil {
					return
				}
			}
			if err = writer.Close(); err != nil {
				return
			}
		case <-ticker.C:
			err = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if err = c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs - handles websocket requests from the peer.
func ServeWs(hub *hub) gin.HandlerFunc {
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
		client.hub.register <- client
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
