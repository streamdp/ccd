package kraken

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_convertKrakenRestDataToDomain(t *testing.T) {
	type args struct {
		from       string
		to         string
		d          *restData
		lastUpdate int64
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
				d: &restData{
					Result: map[string]restTickerInfo{},
				},
				lastUpdate: 0,
			},
			want: nil,
		},
		{
			name: "nil",
			args: args{
				from:       "BTC",
				to:         "USDT",
				d:          nil,
				lastUpdate: 0,
			},
			want: nil,
		},
		{
			name: "regular conversion",
			args: args{
				from: "XRP",
				to:   "USDT",
				d: &restData{
					Error: nil,
					Result: map[string]restTickerInfo{
						"XRPUSDT": {
							A: []string{"2.17935000", "459", "459.000"},
							B: []string{"2.17896000", "76", "76.000"},
							C: []string{"2.18000000", "13.57465600"},
							V: []string{"139560.82655307", "314445.90752978"},
							P: []string{"2.16724243", "2.17213922"},
							T: []int{769, 1599},
							L: []string{"2.13406000", "2.13406000"},
							H: []string{"2.19821000", "2.20396000"},
							O: "2.15556000",
						},
					},
				},
				lastUpdate: 1746439219990,
			},
			want: &domain.Data{
				FromSymbol:   "XRP",
				ToSymbol:     "USDT",
				Open24Hour:   2.15556,
				Volume24Hour: 314445.90752978,
				Low24Hour:    2.13406,
				High24Hour:   2.20396,
				Price:        2.16724243,
				Supply:       1599,
				LastUpdate:   1746439219990,
				DisplayDataRaw: "{\"from_symbol\":\"XRP\",\"to_symbol\":\"USDT\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":2.15556,\"volume_24_hour\":314445.90752978," +
					"\"volume_24_hour_to\":0,\"low_24_hour\":2.13406,\"high_24_hour\":2.20396,\"price\":2.16724243," +
					"\"supply\":1599,\"mkt_cap\":0,\"last_update\":1746439219990}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertRestDataToDomain(tt.args.from, tt.args.to, tt.args.d, tt.args.lastUpdate)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertRestDataToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_krakenRest_buildURL(t *testing.T) {
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
				u, _ := url.Parse("https://api.kraken.com/0/public/Ticker?pair=btcusdt")

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
				u, _ := url.Parse("https://api.kraken.com/0/public/Ticker?pair=ethusd")

				return u
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &rest{}
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
