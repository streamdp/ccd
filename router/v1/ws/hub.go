package ws

import (
	"encoding/json"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/dbconnectors"
	v1 "github.com/streamdp/ccd/router/v1"
)

type Hub struct {
	clients    map[*Client]bool
	queryQueue chan *Query
	register   chan *Client
	unregister chan *Client
}

type Query struct {
	sender *Client
	query  *v1.PriceQuery
}

func NewHub() *Hub {
	return &Hub{
		queryQueue: make(chan *Query, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run(wc *dataproviders.Workers, db *dbconnectors.Db) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case queue := <-h.queryQueue:
			var binaryString []byte
			if data, err := v1.GetLastPrice(wc, db, queue.query); err == nil {
				if binaryString, err = json.Marshal(&data); err == nil {
					queue.sender.send <- binaryString
				}
			}
		}
	}
}
