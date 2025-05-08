package wsclient

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_huobiWs_pairFromChannelName(t *testing.T) {
	buildChannelNameFunc := func(from, to string) string {
		return fmt.Sprintf("%s/%s", from, to)
	}

	tests := []struct {
		name          string
		subscriptions domain.Subscriptions
		ch            string
		wantFrom      string
		wantTo        string
	}{
		{
			name: "get pair from channel name",
			subscriptions: map[string]*domain.Subscription{
				buildChannelNameFunc("btc", "usdt"): domain.NewSubscription("btc", "usdt", 0),
			},
			ch:       buildChannelNameFunc("btc", "usdt"),
			wantFrom: "BTC",
			wantTo:   "USDT",
		},
		{
			name:          "unknown channel name",
			subscriptions: map[string]*domain.Subscription{},
			ch:            buildChannelNameFunc("btc", "usdt"),
			wantFrom:      "",
			wantTo:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Ws{subscriptions: tt.subscriptions}
			gotFrom, gotTo := h.PairFromChannelName(tt.ch)
			if gotFrom != tt.wantFrom {
				t.Errorf("pairFromChannelName() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
			}
			if gotTo != tt.wantTo {
				t.Errorf("pairFromChannelName() gotTo = %v, want %v", gotTo, tt.wantTo)
			}
		})
	}
}

func TestWs_ListSubscriptions(t *testing.T) {
	buildChannelNameFunc := func(from, to string) string {
		return fmt.Sprintf("%s/%s", from, to)
	}

	tests := []struct {
		name          string
		subscriptions domain.Subscriptions
		want          domain.Subscriptions
	}{
		{
			name: "get list of subscriptions",
			subscriptions: map[string]*domain.Subscription{
				buildChannelNameFunc("btc", "eth"): domain.NewSubscription("btc", "eth", 0),
			},
			want: domain.Subscriptions{
				buildChannelNameFunc("btc", "eth"): domain.NewSubscription("btc", "eth", 0),
			},
		},
		{
			name:          "empty",
			subscriptions: map[string]*domain.Subscription{},
			want:          domain.Subscriptions{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Ws{
				subscriptions:      tt.subscriptions,
				ChannelNameBuilder: buildChannelNameFunc,
			}
			if got := w.ListSubscriptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildWsSessionName(t *testing.T) {
	type args struct {
		from string
		to   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "symbols in the upper case",
			args: args{
				from: "BTC",
				to:   "USDT",
			},
			want: "WS:BTC:USDT",
		},
		{
			name: "symbols in the lower case",
			args: args{
				from: "eth",
				to:   "usdt",
			},
			want: "WS:eth:usdt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildWsSessionName(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("buildWsSessionName() = %v, want %v", got, tt.want)
			}
		})
	}
}
