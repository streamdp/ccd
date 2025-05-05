package kraken

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
			want: "btc/eth",
		},
		{
			name: "usd case",
			args: args{
				from: "btc",
				to:   "usd",
			},
			want: "btc/usd",
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
			h := &ws{
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
			c := &ws{
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
		from       string
		to         string
		ticker     *wsTickerInfo
		lastUpdate int64
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
				ticker: &wsTickerInfo{
					Symbol:    "BTC/USDT",
					Bid:       94183.2,
					BidQty:    17.65613222,
					Ask:       94183.3,
					AskQty:    0.03386559,
					Last:      94183.3,
					Volume:    891.98979302,
					Vwap:      94904.6,
					Low:       93570.1,
					High:      95761.3,
					Change:    -1292.8,
					ChangePct: -1.35,
				},
				lastUpdate: 1719731197640,
			},
			want: &domain.Data{
				Id:              0,
				FromSymbol:      "btc",
				ToSymbol:        "usdt",
				Change24Hour:    -1292.8,
				ChangePct24Hour: -1.35,
				Open24Hour:      0,
				Volume24Hour:    891.98979302,
				Low24Hour:       93570.1,
				High24Hour:      95761.3,
				Price:           94904.6,
				Supply:          0,
				MktCap:          0,
				LastUpdate:      1719731197640,
				DisplayDataRaw: "{\"from_symbol\":\"btc\",\"to_symbol\":\"usdt\",\"change_24_hour\":-1292.8," +
					"\"changepct_24_hour\":-1.35,\"open_24_hour\":0,\"volume_24_hour\":891.98979302," +
					"\"volume_24_hour_to\":0,\"low_24_hour\":93570.1,\"high_24_hour\":95761.3,\"price\":94904.6," +
					"\"supply\":0,\"mkt_cap\":0,\"last_update\":1719731197640}",
			},
		},
		{
			name: "empty",
			args: args{
				ticker: &wsTickerInfo{},
			},
			want: &domain.Data{DisplayDataRaw: "{\"from_symbol\":\"\",\"to_symbol\":\"\",\"change_24_hour\":0," +
				"\"changepct_24_hour\":0,\"open_24_hour\":0,\"volume_24_hour\":0,\"volume_24_hour_to\":0," +
				"\"low_24_hour\":0,\"high_24_hour\":0,\"price\":0,\"supply\":0,\"mkt_cap\":0,\"last_update\":0}"},
		},
		{
			name: "nil",
			args: args{ticker: nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertWsDataToDomain(tt.args.from, tt.args.to, tt.args.ticker, tt.args.lastUpdate)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertHuobiWsDataToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
