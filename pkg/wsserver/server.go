package ws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
	"unsafe"

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

	cancel context.CancelFunc
}

func NewServer(ctx context.Context, l *log.Logger, r clients.RestClient, db db.Database) *Server {
	ctx, cancel := context.WithCancel(ctx)

	server := &Server{
		l:          l,
		clients:    make(map[*client]struct{}),
		clientsMu:  new(sync.RWMutex),
		restClient: r,
		dataBase:   db,

		pipe: make(chan *domain.Data, 1000),

		cancel: cancel,
	}

	go server.gc(ctx)
	go server.processSubscriptions()

	return server
}

func (s *Server) AddClient(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return fmt.Errorf("failed to accept ws handshake: %w", err)
	}

	h := &handler{
		l:             s.l,
		conn:          conn,
		messagePipe:   make(chan []byte, 256),
		rc:            s.restClient,
		db:            s.dataBase,
		subscriptions: cache.New(),
	}
	h.conn.SetReadLimit(maxMessageSize)

	ctx, cancel := context.WithCancel(ctx)

	go h.handleMessagePipe(ctx)
	go h.handleClientRequests(ctx)
	go h.handleHeartbeat(ctx)

	c := &client{
		handler: h,
		cancel:  cancel,
	}
	s.clients[c] = struct{}{}
	s.l.Println(fmt.Sprintf("added new ws client: %v", uintptr(unsafe.Pointer(c))))

	return nil
}

func (s *Server) DataPipe() chan *domain.Data {
	return s.pipe
}

func (s *Server) Close() {
	defer close(s.pipe)
	defer s.cancel()

	s.clientsMu.RLock()
	for c := range s.clients {
		c.cancel()
	}
	s.clientsMu.RUnlock()
}

func (s *Server) processSubscriptions() {
	for data := range s.pipe {
		if len(s.clients) == 0 {
			continue
		}

		subscribers := s.getSubscribers((&pair{
			From: data.FromSymbol,
			To:   data.ToSymbol,
		}).buildName())

		if len(subscribers) == 0 {
			continue
		}

		bytes := (&wsMessage{
			T:         "data",
			Data:      data,
			Timestamp: time.Now().UTC().UnixMilli(),
		}).Bytes()

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

				if err := c.handler.close("inactive client"); err != nil &&
					!errors.As(err, &websocket.CloseError{}) &&
					!errors.Is(err, net.ErrClosed) {
					s.l.Println("failed to close inactive client: " + err.Error())
				}

				s.clientsMu.Lock()
				delete(s.clients, c)
				s.clientsMu.Unlock()

				s.l.Println(fmt.Sprintf("removed ws client: %v", uintptr(unsafe.Pointer(c))))
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
