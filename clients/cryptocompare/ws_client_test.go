package cryptocompare

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_cryptoCompareWs_buildURL(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantU   *url.URL
		wantErr bool
	}{
		{
			name:   "build url",
			apiKey: "4pwP6HmdD0O8PGDAkvKA9DSCGK74Ixma",
			wantU: func() *url.URL {
				u, _ := url.Parse("wss://streamer.cryptocompare.com/v2?api_key=4pwP6HmdD0O8PGDAkvKA9DSCGK74Ixma")

				return u
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ws{apiKey: tt.apiKey}
			gotU, err := c.buildURL()
			if (err != nil) != tt.wantErr {
				t.Errorf("buildURL() error = %v, wantErr %v", err, tt.wantErr)

				return
			}
			if !reflect.DeepEqual(gotU, tt.wantU) {
				t.Errorf("buildURL() gotU = %v, want %v", gotU, tt.wantU)
			}
		})
	}
}

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
				from: "BTC",
				to:   "ETH",
			},
			want: "5~CCCAGG~BTC~ETH",
		},
		{
			name: "build from lower",
			args: args{
				from: "btc",
				to:   "eth",
			},
			want: "5~CCCAGG~BTC~ETH",
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

func Test_convertCryptoCompareWsDataToDomain(t *testing.T) {
	tests := []struct {
		name string
		d    *wsData
		want *domain.Data
	}{
		{
			name: "regular conversion",
			d: &wsData{
				FromSymbol:          "BTC",
				ToSymbol:            "ETH",
				Price:               61391.27,
				LastUpdate:          1719731197640,
				Volume24Hour:        666.1991214442407,
				Volume24HourTo:      40641405.87766658,
				Open24Hour:          60867.47,
				High24Hour:          61403.89,
				Low24Hour:           60867.47,
				CurrentSupply:       30701,
				CurrentSupplyMktCap: 1215020604533,
			},
			want: &domain.Data{
				FromSymbol:   "BTC",
				ToSymbol:     "ETH",
				Open24Hour:   60867.47,
				Volume24Hour: 666.1991214442407,
				Low24Hour:    60867.47,
				High24Hour:   61403.89,
				Price:        61391.27,
				Supply:       30701,
				MktCap:       1215020604533,
				LastUpdate:   1719731197640,
				DisplayDataRaw: "{\"from_symbol\":\"BTC\",\"to_symbol\":\"ETH\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":60867.47,\"volume_24_hour\":666.1991214442407," +
					"\"volume_24_hour_to\":40641405.87766658,\"low_24_hour\":0,\"high_24_hour\":61403.89," +
					"\"price\":61391.27,\"supply\":30701,\"mkt_cap\":1215020604533,\"last_update\":1719731197640}",
			},
		},
		{
			name: "empty",
			d:    &wsData{},
			want: &domain.Data{
				DisplayDataRaw: "{\"from_symbol\":\"\",\"to_symbol\":\"\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":0,\"volume_24_hour\":0,\"volume_24_hour_to\":0," +
					"\"low_24_hour\":0,\"high_24_hour\":0,\"price\":0,\"supply\":0,\"mkt_cap\":0,\"last_update\":0}",
			},
		},
		{
			name: "nil",
			d:    nil,
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertWsDataToDomain(tt.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWsDataToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
