package repos

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolCache_Add(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{
			name: "add btc",
			s:    "BTC",
		},
		{
			name: "add eth",
			s:    "eth",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &SymbolCache{
				c: map[string]struct{}{},
				m: new(sync.RWMutex),
			}
			sc.Add(tt.s)
			assert.True(t, sc.IsPresent(tt.s))
		})
	}
}

func TestSymbolCache_Remove(t *testing.T) {
	type fields struct {
		c map[string]struct{}
	}
	tests := []struct {
		name   string
		fields fields
		s      string
	}{
		{
			name: "remove btc",
			fields: fields{
				c: map[string]struct{}{
					"BTC": {},
					"ETH": {},
					"LTC": {},
				},
			},
			s: "BTC",
		},
		{
			name: "remove eth",
			fields: fields{
				c: map[string]struct{}{
					"BTC": {},
					"ETH": {},
					"LTC": {},
				},
			},
			s: "eth",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &SymbolCache{
				c: tt.fields.c,
				m: new(sync.RWMutex),
			}
			sc.Remove(tt.s)
			assert.False(t, sc.IsPresent(tt.s))
		})
	}
}
