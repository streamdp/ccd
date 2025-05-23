package ws

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/cache"
	v1 "github.com/streamdp/ccd/server/api/v1"
)

func Test_handler_subscribe(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions *cache.Cache
		pairs         []*pair
	}{
		{
			name:          "one pair subscribe",
			subscriptions: cache.New(),
			pairs:         []*pair{{From: "BTC", To: "USDT"}},
		},
		{
			name:          "several pairs subscribe",
			subscriptions: cache.New(),
			pairs: []*pair{
				{From: "BTC", To: "USDT"},
				{From: "ETH", To: "USDT"},
			},
		},
		{
			name:          "subscription already present",
			subscriptions: cacheWithSubscription(&pair{From: "BTC", To: "USDT"}),
			pairs:         []*pair{{From: "BTC", To: "USDT"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				messagePipe:   make(chan []byte, 10),
				subscriptions: tt.subscriptions,
			}
			t.Cleanup(func() { close(h.messagePipe) })

			for i := range tt.pairs {
				h.subscribe(tt.pairs[i])
			}

			subscriptions := h.subscriptions.GetAll()
			if len(subscriptions) != len(tt.pairs) {
				t.Errorf("len(subscription) = %v, want = %v", len(subscriptions), len(tt.pairs))
			}

			for _, p := range tt.pairs {
				if !slices.Contains(subscriptions, p.buildName()) {
					t.Errorf("pair not found in subscriptions slice:  %v", p)
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(func() { cancel() })

			go func() {
				c := 0

				for {
					select {
					case <-ctx.Done():
						t.Error("failed to fetch message from the channel")

						return
					case <-h.messagePipe:
						c++
						if c == len(tt.pairs) {
							cancel()

							return
						}
					}
				}
			}()

			<-ctx.Done()
		})
	}
}

func Test_handler_unsubscribe(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions *cache.Cache
		pairs         []*pair
	}{
		{
			name:          "one pair unsubscribe",
			subscriptions: cacheWithSubscription(&pair{From: "BTC", To: "USDT"}),
			pairs:         []*pair{{From: "BTC", To: "USDT"}},
		},
		{
			name: "several pairs unsubscribe",
			subscriptions: cacheWithSubscription(
				&pair{From: "BTC", To: "USDT"},
				&pair{From: "ETH", To: "USDT"},
			),
			pairs: []*pair{
				{From: "BTC", To: "USDT"},
				{From: "ETH", To: "USDT"},
			},
		},
		{
			name:          "subscription not found",
			subscriptions: cacheWithSubscription(&pair{From: "BTC", To: "USDT"}),
			pairs:         []*pair{{From: "ETH", To: "USDT"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				messagePipe:   make(chan []byte, 10),
				subscriptions: tt.subscriptions,
			}
			t.Cleanup(func() { close(h.messagePipe) })

			for i := range tt.pairs {
				h.unsubscribe(tt.pairs[i])
			}

			subscriptions := h.subscriptions.GetAll()
			for _, p := range tt.pairs {
				if slices.Contains(subscriptions, p.buildName()) {
					t.Errorf("pair found in subscriptions slice:  %v", p)
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(func() { cancel() })

			go func() {
				c := 0

				for {
					select {
					case <-ctx.Done():
						t.Error("fetch message error: timeout exceeded")
					case <-h.messagePipe:
						c++
						if c == len(tt.pairs) {
							cancel()

							return
						}
					}
				}
			}()

			<-ctx.Done()
		})
	}
}

func Test_handler_getLastPrice(t *testing.T) {
	tests := []struct {
		name    string
		rc      clients.RestClient
		db      db.Database
		p       *pair
		want    *domain.Data
		wantErr error
	}{
		{
			name: "get last price from external api",
			rc: &mockRestClient{
				data: &domain.Data{
					FromSymbol: "BTC",
					ToSymbol:   "USDT",
				},
				err: nil,
			},
			db: &mockDatabase{
				dataPipe: make(chan *domain.Data, 1),
			},
			p: &pair{From: "BTC", To: "USDT"},
			want: &domain.Data{
				FromSymbol: "BTC",
				ToSymbol:   "USDT",
			},
			wantErr: nil,
		},
		{
			name: "get last price from database",
			rc: &mockRestClient{
				err: v1.ErrGetPrice,
			},
			db: &mockDatabase{
				data: &domain.Data{
					FromSymbol: "BTC",
					ToSymbol:   "USDT",
				},
				dataPipe: make(chan *domain.Data, 1),
			},
			p: &pair{From: "BTC", To: "USDT"},
			want: &domain.Data{
				FromSymbol: "BTC",
				ToSymbol:   "USDT",
			},
			wantErr: nil,
		},
		{
			name: "failed to get last price",
			rc:   &mockRestClient{err: v1.ErrGetPrice},
			db: &mockDatabase{
				dataPipe: make(chan *domain.Data, 1),
				err:      v1.ErrGetPrice,
			},
			p:       &pair{From: "BTC", To: "USDT"},
			want:    nil,
			wantErr: v1.ErrGetPrice,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				close(tt.db.DataPipe())
			})
			h := &handler{
				rc: tt.rc,
				db: tt.db,
			}
			got, err := h.getLastPrice(tt.p)
			if err != nil && !errors.Is(errors.Unwrap(err), tt.wantErr) {
				t.Errorf("getLastPrice() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLastPrice() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockRestClient struct {
	data *domain.Data
	err  error
}

func (m *mockRestClient) Get(_ string, _ string) (*domain.Data, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.data, nil
}

type mockDatabase struct {
	data     *domain.Data
	dataPipe chan *domain.Data
	err      error
}

func (m *mockDatabase) Insert(_ *domain.Data) (sql.Result, error) {
	return nil, m.err
}

func (m *mockDatabase) GetLast(_ string, _ string) (*domain.Data, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.data, nil
}

func (m *mockDatabase) DataPipe() chan *domain.Data {
	return m.dataPipe
}

func (m *mockDatabase) Close() error {
	return m.err
}
