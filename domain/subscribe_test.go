package domain

import (
	"testing"
)

func TestSubscription_Id(t *testing.T) {
	type fields struct {
		From string
		To   string
		id   int64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "get subscription id=1",
			fields: fields{
				From: "btc",
				To:   "usdt",
				id:   1,
			},
			want: 1,
		},
		{
			name: "get subscription id=2",
			fields: fields{
				From: "eth",
				To:   "eur",
				id:   2,
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Subscription{
				From: tt.fields.From,
				To:   tt.fields.To,
				id:   tt.fields.id,
			}
			if got := s.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}
