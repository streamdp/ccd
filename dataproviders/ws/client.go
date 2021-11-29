package ws

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/streamdp/ccdatacollector/dataproviders/cryptocompare"
	"github.com/streamdp/ccdatacollector/handlers"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
)

type Client struct {
	dataHub *DataHub
	conn    *websocket.Conn
	send    chan []byte
}

func (c *Client) PongHandler(string) (err error) {
	if err = c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}
	return
}

type Subscription struct {
	Action string   `json:"action"`
	Subs   []string `json:"subs"`
}

func (c *Client) Subscribe(list []string) {
	subs := &Subscription{
		Action: "SubAdd",
		Subs:   list,
	}
	binaryString, err := json.Marshal(&subs)
	if err != nil {
		handlers.SystemHandler(err)
	}
	c.send <- binaryString
}

func (c *Client) UnSubscribe(list []string) {
	subs := &Subscription{
		Action: "SubRemove",
		Subs:   list,
	}
	binaryString, err := json.Marshal(&subs)
	if err != nil {
		handlers.SystemHandler(err)
	}
	c.send <- binaryString
}

func (c *Client) readPump() {
	var err error
	var cccAgg *CccAgg
	var heartBeat *HeartBeat
	var buff []byte
	defer func() {
		c.dataHub.unregister <- c
		if err = c.conn.Close(); err != nil {
			return
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	if err = c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	c.conn.SetPongHandler(c.PongHandler)
	for {
		var wsData map[string]interface{}
		if _, buff, err = c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				handlers.SystemHandler(err)
			}
			break
		}
		if err = json.Unmarshal(buff, &wsData); err != nil {
			handlers.SystemHandler(err)
			continue
		}
		switch wsData["TYPE"] {
		case "5":
			if err = json.Unmarshal(buff, &cccAgg); err != nil {
				handlers.SystemHandler(err)
				continue
			}
			c.dataHub.cccAgg <- cccAgg
		case "999":
			if err = json.Unmarshal(buff, &heartBeat); err != nil {
				handlers.SystemHandler(err)
				continue
			}
			c.dataHub.heartBeat <- heartBeat
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

func ServeWs(dataHub *DataHub, dp *cryptocompare.Data) *Client {
	wsUrl, err := url.Parse(dp.GetWsURL())
	if err != nil {
		handlers.SystemHandler(err)
		return nil
	}
	header := http.Header{}
	header.Add("authorization", "Apikey "+dp.GetApiKey())
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl.String(), header)
	if err != nil {
		handlers.SystemHandler(err)
		return nil
	}
	client := &Client{
		dataHub: dataHub,
		conn:    conn,
		send:    make(chan []byte, 256),
	}
	client.dataHub.register <- client
	go client.writePump()
	go client.readPump()
	return client
}
