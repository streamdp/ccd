package huobi

import (
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_buildChannelName(t *testing.T) {
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
			name: "build channel name",
			args: args{
				from: "btc",
				to:   "eth",
			},
			want: "market.btceth.ticker",
		},
		{
			name: "usd case",
			args: args{
				from: "btc",
				to:   "usd",
			},
			want: "market.btcusdt.ticker",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildChannelName(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("buildChannelName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_huobiWs_pairFromChannelName(t *testing.T) {
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
				buildChannelName("btc", "usdt"): domain.NewSubscription("btc", "usdt", 0),
			},
			ch:       buildChannelName("btc", "usdt"),
			wantFrom: "BTC",
			wantTo:   "USDT",
		},
		{
			name:          "unknown channel name",
			subscriptions: map[string]*domain.Subscription{},
			ch:            buildChannelName("btc", "usdt"),
			wantFrom:      "",
			wantTo:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &huobiWs{
				subscriptions: tt.subscriptions,
			}
			gotFrom, gotTo := h.pairFromChannelName(tt.ch)
			if gotFrom != tt.wantFrom {
				t.Errorf("pairFromChannelName() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
			}
			if gotTo != tt.wantTo {
				t.Errorf("pairFromChannelName() gotTo = %v, want %v", gotTo, tt.wantTo)
			}
		})
	}
}

func Test_cryptoCompareWs_ListSubscriptions(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions domain.Subscriptions
		want          domain.Subscriptions
	}{
		{
			name: "get list of subscriptions",
			subscriptions: map[string]*domain.Subscription{
				buildChannelName("btc", "eth"): domain.NewSubscription("btc", "eth", 0),
			},
			want: domain.Subscriptions{
				buildChannelName("btc", "eth"): domain.NewSubscription("btc", "eth", 0),
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
			c := &huobiWs{
				subscriptions: tt.subscriptions,
			}
			if got := c.ListSubscriptions(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertHuobiWsDataToDomain(t *testing.T) {
	type args struct {
		from string
		to   string
		d    *huobiWsData
	}
	tests := []struct {
		name string
		args args
		want *domain.Data
	}{
		{
			name: "regular conversion",
			args: args{
				from: "btc",
				to:   "usdt",
				d: &huobiWsData{
					Ch: "market.btcusdt.ticker",
					Ts: 1719731197640,
					Tick: wsTick{
						Open:   60867.47,
						High:   61403.89,
						Low:    60867.47,
						Amount: 666.1991214442407,
						Vol:    40641405.87766658,
						Count:  30701,
						Bid:    61391.27,
					},
				},
			},
			want: &domain.Data{
				FromSymbol:   "btc",
				ToSymbol:     "usdt",
				Open24Hour:   60867.47,
				Volume24Hour: 666.1991214442407,
				Low24Hour:    60867.47,
				High24Hour:   61403.89,
				Price:        61391.27,
				Supply:       30701,
				LastUpdate:   1719731197640,
				DisplayDataRaw: "{\"from_symbol\":\"btc\",\"to_symbol\":\"usdt\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":60867.47,\"volume_24_hour\":666.1991214442407," +
					"\"volume_24_hour_to\":40641405.87766658,\"low_24_hour\":0,\"high_24_hour\":61403.89," +
					"\"price\":61391.27,\"supply\":30701,\"mkt_cap\":0,\"last_update\":1719731197640}",
			},
		},
		{
			name: "empty",
			args: args{
				d: &huobiWsData{},
			},
			want: &domain.Data{
				DisplayDataRaw: "{\"from_symbol\":\"\",\"to_symbol\":\"\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":0,\"volume_24_hour\":0,\"volume_24_hour_to\":0," +
					"\"low_24_hour\":0,\"high_24_hour\":0,\"price\":0,\"supply\":0,\"mkt_cap\":0,\"last_update\":0}",
			},
		},
		{
			name: "nil",
			args: args{
				d: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertHuobiWsDataToDomain(tt.args.from, tt.args.to, tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertHuobiWsDataToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
