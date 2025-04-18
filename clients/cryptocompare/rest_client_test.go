package cryptocompare

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/streamdp/ccd/domain"
)

func Test_convertToDomain(t *testing.T) {
	type args struct {
		from string
		to   string
		d    *cryptoCompareData
	}
	tests := []struct {
		name string
		args args
		want *domain.Data
	}{
		{
			name: "regular conversion",
			args: args{
				from: "BTC",
				to:   "USDT",
				d: &cryptoCompareData{
					Raw: map[string]map[string]*Response{
						"BTC": {
							"USDT": {
								Change24Hour:    12345,
								Changepct24Hour: 54321,
								Open24Hour:      60867.47,
								Volume24Hour:    666.1991214442407,
								Volume24Hourto:  40641405.87766658,
								Low24Hour:       60867.47,
								High24Hour:      61403.89,
								Price:           61391.27,
								Supply:          30701,
								MktCap:          1215020604533,
								LastUpdate:      1719731197640,
							},
						},
					},
				},
			},
			want: &domain.Data{
				FromSymbol:      "BTC",
				ToSymbol:        "USDT",
				Change24Hour:    12345,
				ChangePct24Hour: 54321,
				Open24Hour:      60867.47,
				Volume24Hour:    666.1991214442407,
				Low24Hour:       60867.47,
				High24Hour:      61403.89,
				Price:           61391.27,
				Supply:          30701,
				MktCap:          1215020604533,
				LastUpdate:      1719731197640,
				DisplayDataRaw: "{\"from_symbol\":\"BTC\",\"to_symbol\":\"USDT\",\"change_24_hour\":0," +
					"\"changepct_24_hour\":0,\"open_24_hour\":60867.47,\"volume_24_hour\":666.1991214442407," +
					"\"volume_24_hour_to\":40641405.87766658,\"low_24_hour\":0,\"high_24_hour\":61403.89," +
					"\"price\":61391.27,\"supply\":30701,\"mkt_cap\":1215020604533,\"last_update\":1719731197640}",
			},
		},
		{
			name: "nil",
			args: args{
				from: "BTC",
				to:   "USDT",
				d: &cryptoCompareData{
					Raw: map[string]map[string]*Response{
						"BTC": nil,
					},
				},
			},
			want: nil,
		},
		{
			name: "empty",
			args: args{
				from: "BTC",
				to:   "USDT",
				d: &cryptoCompareData{
					Raw: map[string]map[string]*Response{"BTC": {"USDT": {}}},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertToDomain(tt.args.from, tt.args.to, tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertToDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cryptoCompareRest_buildURL(t *testing.T) {
	type fields struct {
		apiKey string
		client *http.Client
	}
	type args struct {
		fSym string
		tSym string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantU   *url.URL
		wantErr bool
	}{
		{
			name: "build url",
			fields: fields{
				apiKey: "jHQuBvBisp3UFKzqvmWpH4elAqNv+JQT",
				client: nil,
			},
			args: args{
				fSym: "BTC",
				tSym: "USTD",
			},
			wantU: func() *url.URL {
				u, _ := url.Parse(
					"https://min-api.cryptocompare.com/data/pricemultifull?" +
						"api_key=jHQuBvBisp3UFKzqvmWpH4elAqNv%2BJQT&fsyms=BTC&tsyms=USTD",
				)

				return u
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &cryptoCompareRest{
				apiKey: tt.fields.apiKey,
				client: tt.fields.client,
			}
			gotU, err := cc.buildURL(tt.args.fSym, tt.args.tSym)
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
