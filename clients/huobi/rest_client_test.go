package huobi

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_convertHuobiRestDataToDomain(t *testing.T) {
	type args struct {
		from string
		to   string
		d    *huobiRestData
	}
	tests := []struct {
		name string
		args args
		want *domain.Data
	}{
		{
			name: "empty",
			args: args{
				from: "BTC",
				to:   "USDT",
				d: &huobiRestData{
					Tick: restTick{},
				},
			},
			want: &domain.Data{
				FromSymbol: "BTC",
				ToSymbol:   "USDT",
				DisplayDataRaw: "{\"from_symbol\":\"BTC\",\"to_symbol\":\"USDT\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":0,\"volume_24_hour\":0,\"volume_24_hour_to\":0," +
					"\"low_24_hour\":0,\"high_24_hour\":0,\"price\":0,\"supply\":0,\"mkt_cap\":0,\"last_update\":0}",
			},
		},
		{
			name: "nil",
			args: args{
				from: "BTC",
				to:   "USDT",
				d:    nil,
			},
			want: nil,
		},
		{
			name: "regular conversion",
			args: args{
				from: "BTC",
				to:   "USDT",
				d: &huobiRestData{
					Ts: 1719731197640,
					Tick: restTick{
						Open:   60867.47,
						Low:    60867.47,
						High:   61403.89,
						Amount: 666.1991214442407,
						Vol:    40641405.87766658,
						Count:  30701,
						Bid:    []float64{61391.27},
					},
				},
			},
			want: &domain.Data{
				FromSymbol:   "BTC",
				ToSymbol:     "USDT",
				Open24Hour:   60867.47,
				Volume24Hour: 666.1991214442407,
				Low24Hour:    60867.47,
				High24Hour:   61403.89,
				Price:        61391.27,
				Supply:       30701,
				LastUpdate:   1719731197640,
				DisplayDataRaw: "{\"from_symbol\":\"BTC\",\"to_symbol\":\"USDT\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":60867.47,\"volume_24_hour\":666.1991214442407," +
					"\"volume_24_hour_to\":40641405.87766658,\"low_24_hour\":0,\"high_24_hour\":61403.89," +
					"\"price\":61391.27,\"supply\":30701,\"mkt_cap\":0,\"last_update\":1719731197640}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertHuobiRestDataToDomain(tt.args.from, tt.args.to, tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertHuobiRestDataToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_huobiRest_buildURL(t *testing.T) {
	type args struct {
		fSym string
		tSym string
	}
	tests := []struct {
		name    string
		args    args
		wantU   *url.URL
		wantErr bool
	}{
		{
			name: "build url",
			args: args{
				fSym: "BTC",
				tSym: "USDT",
			},
			wantU: func() *url.URL {
				u, _ := url.Parse("https://api.huobi.pro/market/detail/merged?symbol=btcusdt")

				return u
			}(),
			wantErr: false,
		},
		{
			name: "usd case",
			args: args{
				fSym: "ETH",
				tSym: "USD",
			},
			wantU: func() *url.URL {
				u, _ := url.Parse("https://api.huobi.pro/market/detail/merged?symbol=ethusdt")

				return u
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &huobiRest{}
			gotU, err := h.buildURL(tt.args.fSym, tt.args.tSym)
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
