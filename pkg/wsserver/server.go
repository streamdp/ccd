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

type Server struct {
	l *log.Logger

	clients   map[*client]struct{}
	clientsMu *sync.RWMutex

	restClient clients.RestClient
	dataBase   db.Database

	pipe chan *domain.Data
}

type client struct {
	handler *wsHandler
	cancel  context.CancelFunc
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

	h.sendWelcomeMessage()

	s.clients[&client{
		handler: h,
		cancel:  cancel,
	}] = struct{}{}

	return nil
}

func (s *Server) processSubscriptions() {
	for data := range s.pipe {
		clientData := data.Marshal()

		p := &pair{
			From: data.FromSymbol,
			To:   data.ToSymbol,
		}
		subscriptionName := p.buildName()

		s.clientsMu.RLock()
		for c := range s.clients {
			if c.handler.isActive && c.handler.subscriptions.IsPresent(subscriptionName) {
				c.handler.messagePipe <- clientData
			}
		}
		s.clientsMu.RUnlock()
	}
}

func (s *Server) gc(ctx context.Context) {
	t := time.NewTimer(10 * time.Second)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if len(s.clients) == 0 {
				continue
			}

			s.clientsMu.Lock()
			for k := range s.clients {
				if !k.handler.isActive {
					delete(s.clients, k)
				}
			}
			s.clientsMu.Unlock()
		}
	}
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
