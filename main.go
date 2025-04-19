package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/clients/cryptocompare"
	"github.com/streamdp/ccd/clients/huobi"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/db/redis"
	"github.com/streamdp/ccd/repos"
	"github.com/streamdp/ccd/router"
)

var (
	errInitRestClient   = errors.New("failed to initialize rest client")
	errInitWsClient     = errors.New("failed to initialize ws client")
	errInitSessionStore = errors.New("failed to init session store")
)

func main() {
	l := log.New(gin.DefaultWriter, "CCD:", log.LstdFlags)

	config.ParseFlags()
	gin.SetMode(config.RunMode)

	d, err := db.Connect()
	if err != nil {
		l.Fatalln(err)
	}
	database, ok := d.(db.Database)
	if !ok {
		l.Fatalln("database type assertion error")
	}
	defer func() {
		if errClose := database.Close(); errClose != nil {
			l.Printf("failed to close database connection: %v", errClose)
		}
	}()
	go db.Serve(database, l)

	taskRepo, ok := d.(repos.TaskStore)
	if !ok {
		l.Fatalln("task repo type assertion error")
	}
	s, err := newSessionStore(taskRepo)
	if err != nil {
		l.Fatal(err)
	}
	defer func() {
		if errClose := s.Close(); errClose != nil {
			l.Printf("failed to close session store: %v", errClose)
		}
	}()

	symbolRepo, ok := d.(repos.SymbolsStore)
	if !ok {
		l.Fatalln("symbol repo type assertion error")
	}
	sr := repos.NewSymbolRepository(symbolRepo)
	if err = sr.Load(); err != nil {
		l.Fatalln(err)
	}

	r, err := initRestClient()
	if err != nil {
		l.Fatalln(err)
	}

	ctx := context.Background()

	w, err := initWsClient(ctx, database, l)
	if err != nil {
		l.Fatalln(err)
	}

	p := clients.NewPuller(r, l, s, database.DataPipe())
	if err = p.RestoreLastSession(); err != nil {
		l.Printf("error restoring last session: %v", err)
	}

	e := gin.Default()
	if err = router.InitRouter(ctx, e, database, l, sr, r, w, p); err != nil {
		l.Fatalln(err)
	}

	if err = e.Run(config.Port); err != nil {
		l.Fatalln(err)
	}

	<-ctx.Done()
}

func initRestClient() (clients.RestClient, error) {
	var (
		r   clients.RestClient
		err error
	)

	switch config.DataProvider {
	case "huobi":
		r, err = huobi.Init()
	default:
		r, err = cryptocompare.Init()
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitRestClient, err)
	}

	return r, nil
}

func initWsClient(ctx context.Context, d db.Database, l *log.Logger) (clients.WsClient, error) {
	var (
		w   clients.WsClient
		err error
	)

	switch config.DataProvider {
	case "huobi":
		w, err = huobi.InitWs(ctx, d.DataPipe(), l)
	default:
		w, err = cryptocompare.InitWs(ctx, d.DataPipe(), l)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitWsClient, err)
	}

	return w, nil
}

func newSessionStore(t repos.TaskStore) (db.Session, error) {
	var (
		s   db.Session
		err error
	)

	switch config.SessionStore {
	case "redis":
		s, err = redis.NewRedisKeysStore()
	default:
		s, err = repos.NewSessionRepo(t)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitSessionStore, err)
	}

	return s, nil
}
