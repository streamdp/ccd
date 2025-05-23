package ws

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func cacheWithSubscription(p ...*pair) *cache.Cache {
	c := cache.New()
	for i := range p {
		c.Add(p[i].buildName())
	}

	return c
}

func TestServer_getInactiveClients(t *testing.T) {
	tests := []struct {
		name    string
		clients map[*client]struct{}
		want    []*client
	}{
		{
			name: "get inactive clients",
			clients: map[*client]struct{}{
				&client{handler: &handler{isActive: false}}: {},
				&client{handler: &handler{isActive: true}}:  {},
				&client{handler: &handler{isActive: false}}: {},
				&client{handler: &handler{isActive: true}}:  {},
			},
			want: []*client{
				{handler: &handler{isActive: false}},
				{handler: &handler{isActive: false}},
			},
		},
		{
			name: "no inactive clients found",
			clients: map[*client]struct{}{
				&client{handler: &handler{isActive: true}}: {},
				&client{handler: &handler{isActive: true}}: {},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				clients:   tt.clients,
				clientsMu: new(sync.RWMutex),
			}
			if got := s.getInactiveClients(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getInactiveClients() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_getSubscribers(t *testing.T) {
	tests := []struct {
		name         string
		clients      map[*client]struct{}
		subscription string
		want         []*client
	}{
		{
			name: "get one active clients",
			clients: map[*client]struct{}{
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
						),
						isActive: true,
					},
				}: {},
				&client{handler: &handler{isActive: false}}: {},
			},
			subscription: (&pair{From: "BTC", To: "USDT"}).buildName(),
			want: []*client{{
				handler: &handler{
					subscriptions: cacheWithSubscription(
						&pair{From: "BTC", To: "USDT"},
					),
					isActive: true,
				},
			}},
		},
		{
			name: "get several active clients",
			clients: map[*client]struct{}{
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				}: {},
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "LTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				}: {},
				&client{handler: &handler{isActive: false}}: {},
			},
			subscription: (&pair{From: "ETH", To: "USDT"}).buildName(),
			want: []*client{
				{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				},
				{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "LTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				},
			},
		},
		{
			name: "there are no active clients",
			clients: map[*client]struct{}{
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				}: {},
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "LTC", To: "USDT"},
							&pair{From: "ETH", To: "USDT"},
						),
						isActive: true,
					},
				}: {},
				&client{handler: &handler{isActive: false}}: {},
			},
			subscription: (&pair{From: "XRP", To: "USDT"}).buildName(),
			want:         nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				clients:   tt.clients,
				clientsMu: new(sync.RWMutex),
			}
			got := s.getSubscribers(tt.subscription)
			assert.ElementsMatchf(t, got, tt.want, "getSubscribers() = %v, want %v", got, tt.want)
		})
	}
}

func TestServer_processSubscriptions(t *testing.T) {
	tests := []struct {
		name    string
		clients map[*client]struct{}
		data    *domain.Data
	}{
		{
			name: "one client active client",
			clients: map[*client]struct{}{
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
						),
						messagePipe: make(chan []byte, 10),
						isActive:    true,
					},
				}: {},
				&client{handler: &handler{isActive: false}}: {},
			},
			data: &domain.Data{
				FromSymbol: "BTC",
				ToSymbol:   "USDT",
			},
		},
		{
			name: "send data to the several clients",
			clients: map[*client]struct{}{
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "BTC", To: "USDT"},
						),
						messagePipe: make(chan []byte, 10),
						isActive:    true,
					},
				}: {},
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "ETH", To: "USDT"},
						),
						messagePipe: make(chan []byte, 10),
						isActive:    true,
					},
				}: {},
				&client{
					handler: &handler{
						subscriptions: cacheWithSubscription(
							&pair{From: "LTC", To: "USDT"},
							&pair{From: "BTC", To: "USDT"},
						),
						messagePipe: make(chan []byte, 10),
						isActive:    true,
					},
				}: {},
				&client{handler: &handler{isActive: false}}: {},
			},
			data: &domain.Data{
				FromSymbol: "BTC",
				ToSymbol:   "USDT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				clients:   tt.clients,
				clientsMu: new(sync.RWMutex),
				pipe:      make(chan *domain.Data, 10),
			}
			t.Cleanup(func() { close(s.pipe) })

			go s.processSubscriptions()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(func() { cancel() })

			wg := sync.WaitGroup{}
			for _, c := range s.getSubscribers((&pair{From: tt.data.FromSymbol, To: tt.data.ToSymbol}).buildName()) {
				if !c.handler.isActive {
					continue
				}

				wg.Add(1)
				go func() {
					defer wg.Done()
					defer close(c.handler.messagePipe)

					for {
						select {
						case <-ctx.Done():
							t.Errorf("failed to fetch message: timout exceeded")

							return
						case msg := <-c.handler.messagePipe:
							wsMsg := wsMessage{}
							if err := json.Unmarshal(msg, &wsMsg); err != nil {
								t.Errorf("failed to unmarshal message: %v", err)
							}
							if wsMsg.T != messageTypeData {
								t.Errorf("wrong message type: got = %v, want = %v", wsMsg.T, messageTypeData)
							}
							if !reflect.DeepEqual(wsMsg.Data, tt.data) {
								t.Errorf("wrong message data: got = %v, want = %v", wsMsg.Data, tt.data)
							}

							return
						}
					}
				}()
			}

			s.pipe <- tt.data
			wg.Wait()
		})
	}
}
