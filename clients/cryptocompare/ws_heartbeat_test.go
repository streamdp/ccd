package cryptocompare

import (
	"testing"
)

func Test_heartbeat_reset(t *testing.T) {
	tests := []struct {
		name string
		hb   *heartbeat
		want int64
	}{
		{
			name: "set counter = -1",
			hb:   &heartbeat{c: -1},
			want: heartbeatInitCounter,
		},
		{
			name: "set counter = 10",
			hb:   &heartbeat{c: 10},
			want: heartbeatInitCounter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.hb.reset()
			if tt.hb.c != tt.want {
				t.Fatalf("reset(): got = %d, want = %d", tt.hb.c, tt.want)
			}
		})
	}
}

func Test_heartbeat_decrease(t *testing.T) {
	tests := []struct {
		name string
		hb   *heartbeat
		want int64
	}{
		{
			name: "decrease counter from 10 to 9",
			hb:   &heartbeat{c: 10},
			want: 9,
		},
		{
			name: "decrease counter from -1 to -2",
			hb:   &heartbeat{c: -1},
			want: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.hb.decrease()
			if tt.hb.c != tt.want {
				t.Fatalf("decrease(): got = %d, want = %d", tt.hb.c, tt.want)
			}
		})
	}
}

func Test_heartbeat_isLost(t *testing.T) {
	tests := []struct {
		name string
		hb   *heartbeat
		want bool
	}{
		{
			name: "heartbeat is not lost",
			hb:   newHeartbeat(),
			want: false,
		},
		{
			name: "heartbeat lost because less than 0",
			hb:   &heartbeat{c: -2},
			want: true,
		},
		{
			name: "heartbeat lost because equal 0",
			hb:   &heartbeat{c: 0},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hb.isLost(); got != tt.want {
				t.Errorf("isLost() = %v, want %v", got, tt.want)
			}
		})
	}
}
