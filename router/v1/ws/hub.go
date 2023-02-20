package ws

import (
	"encoding/json"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/dbconnectors"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
)

type Hub struct {
	clients    map[*Client]struct{}
	queryQueue chan *Query
	unregister chan *Client
}

// Query basic structure for build query queue
type Query struct {
	send  chan []byte
	query *v1.PriceQuery
}

// NewHub init new hub
func NewHub() *Hub {
	return &Hub{
		queryQueue: make(chan *Query, 256),
		unregister: make(chan *Client),
		clients:    make(map[*Client]struct{}),
	}
}

// Run loop what serving register/unregister and queryQueue chan
func (h *Hub) Run(wc *dataproviders.Workers, db *dbconnectors.Db) {
	go func() {
		for {
			select {
			case client := <-h.unregister:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			case queue := <-h.queryQueue:
				var (
					data         *dataproviders.Data
					binaryString []byte
					err          error
				)
				if data, err = v1.GetLastPrice(wc, db, queue.query); err != nil {
					handlers.SystemHandler(err)
					continue
				}
				if binaryString, err = json.Marshal(&data); err != nil {
					handlers.SystemHandler(err)
					continue
				}
				queue.send <- binaryString
			}
		}
	}()
}
