package clients

import (
	"context"

	"github.com/streamdp/ccd/domain"
)

// RestClient interface makes it possible to expand the list of rest data providers
type RestClient interface {
	Get(from string, to string) (*domain.Data, error)
}

// WsClient interface makes it possible to expand the list of wss data providers
type WsClient interface {
	Subscribe(ctx context.Context, from string, to string) error
	Unsubscribe(ctx context.Context, from string, to string) error
	ListSubscriptions() domain.Subscriptions
}
