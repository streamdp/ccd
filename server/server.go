package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	v1 "github.com/streamdp/ccd/server/api/v1"
)

type Server struct {
	*gin.Engine

	d  db.Database
	sr v1.SymbolsRepo
	r  clients.RestClient
	w  clients.WsClient
	p  clients.RestApiPuller

	l   *log.Logger
	cfg *config.App
}

func NewServer(
	d db.Database,
	sr v1.SymbolsRepo,
	r clients.RestClient,
	w clients.WsClient,
	p clients.RestApiPuller,
	l *log.Logger,
	cfg *config.App,
) *Server {
	return &Server{
		Engine: gin.Default(),

		d:  d,
		sr: sr,
		r:  r,
		w:  w,
		p:  p,

		l:   l,
		cfg: cfg,
	}
}
