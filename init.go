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
	"github.com/streamdp/ccd/pkg/sessionrepo"
)

var (
	errInitRestClient   = errors.New("failed to initialize rest client")
	errInitWsClient     = errors.New("failed to initialize ws client")
	errInitSessionStore = errors.New("failed to init session store")
)

func initRestClient(cfg *config.App) (clients.RestClient, error) {
	var (
		restClient clients.RestClient
		err        error
	)

	switch cfg.DataProvider {
	case "huobi":
		restClient, err = huobi.Init(cfg)
	case "kraken":
		restClient, err = kraken.Init(cfg)
	default:
		restClient, err = cryptocompare.Init(cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitRestClient, err)
	}

	return restClient, nil
}

func initWsClient(
	ctx context.Context, d db.Database, sessionRepo clients.SessionRepo, l *log.Logger, cfg *config.App,
) (clients.WsClient, error) {
	var (
		wsClient clients.WsClient
		err      error
	)

	switch cfg.DataProvider {
	case "huobi":
		wsClient = huobi.InitWs(ctx, d.DataPipe(), sessionRepo, l, cfg.Http)
	case "kraken":
		wsClient = kraken.InitWs(ctx, d.DataPipe(), sessionRepo, l, cfg.Http)
	default:
		wsClient, err = cryptocompare.InitWs(ctx, d.DataPipe(), sessionRepo, l, cfg.Http, cfg.ApiKey)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitWsClient, err)
	}

	return wsClient, nil
}

func newSessionRepo(s sessionrepo.SessionStore, cfg *config.App) (clients.SessionRepo, error) {
	var (
		sessionRepo clients.SessionRepo
		err         error
	)

	switch cfg.SessionStore {
	case "redis":
		sessionRepo, err = redis.NewRedisKeysStore(cfg)
	default:
		sessionRepo, err = sessionrepo.New(s)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInitSessionStore, err)
	}

	return sessionRepo, nil
}
