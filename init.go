package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/clients/cryptocompare"
	"github.com/streamdp/ccd/clients/huobi"
	"github.com/streamdp/ccd/clients/kraken"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/db/redis"
	"github.com/streamdp/ccd/repos"
)

var (
	errInitRestClient   = errors.New("failed to initialize rest client")
	errInitWsClient     = errors.New("failed to initialize ws client")
	errInitSessionStore = errors.New("failed to init session store")
)

func initRestClient(cfg *config.App) (clients.RestClient, error) {
	var (
		r   clients.RestClient
		err error
	)

	switch cfg.DataProvider {
	case "huobi":
		r, err = huobi.Init(cfg)
	case "kraken":
		r, err = kraken.Init(cfg)
	default:
		r, err = cryptocompare.Init(cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitRestClient, err)
	}

	return r, nil
}

func initWsClient(ctx context.Context, d db.Database, l *log.Logger, cfg *config.App) (clients.WsClient, error) {
	var (
		w   clients.WsClient
		err error
	)

	switch cfg.DataProvider {
	case "huobi":
		w, err = huobi.InitWs(ctx, d.DataPipe(), l)
	case "kraken":
		w, err = kraken.InitWs(ctx, d.DataPipe(), l)
	default:
		w, err = cryptocompare.InitWs(ctx, d.DataPipe(), l, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitWsClient, err)
	}

	return w, nil
}

func newSessionStore(t repos.TaskStore, cfg *config.App) (clients.Session, error) {
	var (
		s   clients.Session
		err error
	)

	switch cfg.SessionStore {
	case "redis":
		s, err = redis.NewRedisKeysStore(cfg)
	default:
		s, err = repos.NewSessionRepo(t)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitSessionStore, err)
	}

	return s, nil
}
