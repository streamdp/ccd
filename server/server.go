package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	ws "github.com/streamdp/ccd/pkg/wsserver"
	v1 "github.com/streamdp/ccd/server/api/v1"
)

type server struct {
	*gin.Engine

	d  db.Database
	sr v1.SymbolsRepo
	rc clients.RestClient
	wc clients.WsClient
	p  v1.Puller

	l   *log.Logger
	cfg *config.App

	ws *ws.Server
}

func NewServer(
	d db.Database,
	sr v1.SymbolsRepo,
	rc clients.RestClient,
	wc clients.WsClient,
	p v1.Puller,
	l *log.Logger,
	cfg *config.App,
	ws *ws.Server,
) *server {
	return &server{
		Engine: gin.Default(),

		d:  d,
		sr: sr,
		rc: rc,
		wc: wc,
		p:  p,

		l:   l,
		cfg: cfg,

		ws: ws,
	}
}

func (s *server) Run(addr string, serverTimeout time.Duration) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s,
		IdleTimeout:       serverTimeout,
		ReadTimeout:       serverTimeout,
		ReadHeaderTimeout: serverTimeout,
		WriteTimeout:      serverTimeout,
	}

	log.Fatalln(srv.ListenAndServe())
}
