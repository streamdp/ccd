package ws

import (
	"encoding/json"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/dbconnectors"
	v1 "github.com/streamdp/ccd/router/v1"
)

type hub struct {
	clients    map[*Client]bool
	queryQueue chan *Query
	register   chan *Client
	unregister chan *Client
}

// Query basic structure for build query queue
type Query struct {
	sender *Client
	query  *v1.PriceQuery
}

// NewHub init new hub
func NewHub() *hub {
	return &hub{
		queryQueue: make(chan *Query, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run loop what serving register/unregister and queryQueue chan
func (h *hub) Run(wc *dataproviders.Workers, db *dbconnectors.Db) {
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
