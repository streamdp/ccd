package kraken

import (
	"time"
)

type restData struct {
	Error  []interface{} `json:"error"`
	Result map[string]restTickerInfo
}

type restTickerInfo struct {
	A []string `json:"a"`
	B []string `json:"b"`
	C []string `json:"c"`
	V []string `json:"v"`
	P []string `json:"p"`
	T []int    `json:"t"`
	L []string `json:"l"`
	H []string `json:"h"`
	O string   `json:"o"`
}

type wsMessage struct {
	Method string           `json:"method"`
	Params *wsMessageParams `json:"params,omitempty"`
	Result *struct {
		Channel  string `json:"channel"`
		Snapshot bool   `json:"snapshot"`
		Symbol   string `json:"symbol"`
	} `json:"result,omitempty"`
	Success bool       `json:"success,omitempty"`
	Error   string     `json:"error,omitempty"`
	TimeIn  *time.Time `json:"time_in,omitempty"`
	TimeOut *time.Time `json:"time_out,omitempty"`
}

type wsMessageParams struct {
	Channel string   `json:"channel"`
	Symbol  []string `json:"symbol"`
}

type wsData struct {
	Channel string `json:"channel"`
	Type    string `json:"type"`
	Data    []wsTickerInfo
}

type wsTickerInfo struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	BidQty    float64 `json:"bid_qty"`
	Ask       float64 `json:"ask"`
	AskQty    float64 `json:"ask_qty"`
	Last      float64 `json:"last"`
	Volume    float64 `json:"volume"`
	Vwap      float64 `json:"vwap"`
	Low       float64 `json:"low"`
	High      float64 `json:"high"`
	Change    float64 `json:"change"`
	ChangePct float64 `json:"change_pct"`
}
