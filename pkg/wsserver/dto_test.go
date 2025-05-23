package ws

import (
	"reflect"
	"testing"
)

func Test_pair_buildName(t *testing.T) {
	tests := []struct {
		name string
		p    *pair
		want string
	}{
		{
			name: "get btc/usdt pair name",
			p: &pair{
				From: "BTC",
				To:   "USDT",
			},
			want: "BTC:USDT",
		},
		{
			name: "get eth/usdt pair name",
			p: &pair{
				From: "eth",
				To:   "usdt",
			},
			want: "eth:usdt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.buildName(); got != tt.want {
				t.Errorf("buildName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pair_toUpper(t *testing.T) {
	tests := []struct {
		name string
		p    *pair
		want *pair
	}{
		{
			name: "lower case input",
			p: &pair{
				From: "btc",
				To:   "usdt",
			},
			want: &pair{
				From: "BTC",
				To:   "USDT",
			},
		},
		{
			name: "mixed input",
			p: &pair{
				From: "ETH",
				To:   "usdt",
			},
			want: &pair{
				From: "ETH",
				To:   "USDT",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.toUpper()
			if !reflect.DeepEqual(tt.p, tt.want) {
				t.Errorf("toUpper() gotU = %v, want %v", tt.p, tt.want)
			}
		})
	}
}
