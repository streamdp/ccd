package clients

import (
	"context"

	"github.com/streamdp/ccd/domain"
)

type RestClient interface {
	Get(from string, to string) (*domain.Data, error)
}

type WsClient interface {
	Subscribe(ctx context.Context, from string, to string) error
	Unsubscribe(ctx context.Context, from string, to string) error
	ListSubscriptions() domain.Subscriptions
	RestoreLastSession(ctx context.Context) error
}

type SessionRepo interface {
	AddTask(ctx context.Context, n string, i int64) (err error)
	UpdateTask(ctx context.Context, n string, i int64) (err error)
	RemoveTask(ctx context.Context, n string) (err error)
	GetSession(ctx context.Context) (map[string]int64, error)

	Close() error
}
