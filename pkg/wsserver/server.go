package ws

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/cache"
)

const gcInterval = 30 * time.Second

type Server struct {
	l *log.Logger

	clients   map[*client]struct{}
	clientsMu *sync.RWMutex

	restClient clients.RestClient
	dataBase   db.Database

	pipe chan *domain.Data
}

func NewServer(ctx context.Context, r clients.RestClient, l *log.Logger, db db.Database) *Server {
	server := &Server{
		l:          l,
		clients:    make(map[*client]struct{}),
		clientsMu:  new(sync.RWMutex),
		restClient: r,
		dataBase:   db,

		pipe: make(chan *domain.Data, 1000),
	}

	go server.gc(ctx)
	go server.processSubscriptions()

	return server
}

func (s *Server) AddClient(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return err
	}

	h := &wsHandler{
		l:             s.l,
		conn:          conn,
		messagePipe:   make(chan []byte, 256),
		rc:            s.restClient,
		db:            s.dataBase,
		subscriptions: cache.New(),
		isActive:      true,
	}
	h.conn.SetReadLimit(maxMessageSize)

	ctx, cancel := context.WithCancel(ctx)

	go h.handleMessagePipe(ctx)
	go h.handleClientRequests(ctx)

	s.clients[&client{
		handler: h,
		cancel:  cancel,
	}] = struct{}{}

	return nil
}

func (s *Server) DataPipe() chan *domain.Data {
	return s.pipe
}

func (s *Server) Close() {
	s.clientsMu.RLock()
	for c := range s.clients {
		c.cancel()
	}
	s.clientsMu.RUnlock()
}

func (s *Server) processSubscriptions() {
	for data := range s.pipe {
		subscribers := s.getSubscribers((&pair{
			From: data.FromSymbol,
			To:   data.ToSymbol,
		}).buildName())

		if len(subscribers) == 0 {
			continue
		}

		bytes := (&wsMessage{
			T:    "data",
			Data: data,
		}).Marshal()

		for _, c := range subscribers {
			c.handler.messagePipe <- bytes
		}
	}
}

func (s *Server) getSubscribers(subscription string) []*client {
	var res []*client

	s.clientsMu.RLock()
	for c := range s.clients {
		if c.handler.isActive && c.handler.subscriptions.IsPresent(subscription) {
			res = append(res, c)
		}
	}
	s.clientsMu.RUnlock()

	return res
}

func (s *Server) gc(ctx context.Context) {
	t := time.NewTimer(gcInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			t.Reset(gcInterval)

			if len(s.clients) == 0 {
				continue
			}

			for _, c := range s.getInactiveClients() {
				c.cancel()

				if err := c.handler.conn.Ping(ctx); err == nil {
					if err = c.handler.Close("inactive client"); err != nil {
						s.l.Println("failed to close inactive client: " + err.Error())
					}
				}

				s.clientsMu.Lock()
				delete(s.clients, c)
				s.clientsMu.Unlock()
			}
		}
	}
}

func (s *Server) getInactiveClients() []*client {
	var res []*client

	s.clientsMu.RLock()
	for c := range s.clients {
		if !c.handler.isActive {
			res = append(res, c)
		}
	}
	s.clientsMu.RUnlock()

	return res
}
