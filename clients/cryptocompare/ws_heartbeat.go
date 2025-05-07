package cryptocompare

import (
	"errors"
	"time"
)

// CryptoCompare automatically send a heartbeat message per socket every 30 seconds,
// if we miss two heartbeats it means our connection might be stale. So let's set hb=2 and
// add reset every time the server sends a heartbeat and subtract 1 by ticker every 30 seconds.
const (
	heartbeatInitCounter   = 2
	heartbeatCheckInterval = 30 * time.Second
)

type heartbeat struct {
	c int64
}

var errHeartbeat = errors.New("heartbeat loss")

func newHeartbeat() *heartbeat {
	return &heartbeat{c: heartbeatInitCounter}
}

func (hb *heartbeat) reset() {
	hb.c = heartbeatInitCounter
}

func (hb *heartbeat) decrease() {
	hb.c--
}

func (hb *heartbeat) isLost() bool {
	return hb.c <= 0
}
