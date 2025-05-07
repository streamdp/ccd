package huobi

import (
	"encoding/json"
)

type restData struct {
	Ch      string   `json:"ch"`
	Status  string   `json:"status"`
	ErrCode string   `json:"err-code"`
	ErrMsg  string   `json:"err-msg"`
	Ts      int64    `json:"ts"`
	Tick    restTick `json:"tick"`
}

type restTick struct {
	Id      int64     `json:"id"`
	Version int64     `json:"version"`
	Open    float64   `json:"open"`
	Close   float64   `json:"close"`
	Low     float64   `json:"low"`
	High    float64   `json:"high"`
	Amount  float64   `json:"amount"`
	Vol     float64   `json:"vol"`
	Count   int       `json:"count"`
	Bid     []float64 `json:"bid"`
	Ask     []float64 `json:"ask"`
}

type wsMessage struct {
	Id       string `json:"id"`
	Status   string `json:"status"`
	Subbed   string `json:"subbed,omitempty"`
	Unsubbed string `json:"unsubbed,omitempty"`

	Ts int64 `json:"ts"`
}

type wsData struct {
	Ch   string `json:"ch"`
	Ts   int64  `json:"ts"`
	Tick wsTick `json:"tick"`
}

type wsTick struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Amount    float64 `json:"amount"`
	Vol       float64 `json:"vol"`
	Count     int     `json:"count"`
	Bid       float64 `json:"bid"`
	BidSize   float64 `json:"bidSize"`
	Ask       float64 `json:"ask"`
	AskSize   float64 `json:"askSize"`
	LastPrice float64 `json:"lastPrice"`
	LastSize  float64 `json:"lastSize"`
}
type wsPing struct {
	Ts int64 `json:"ping"`
}

func (p *wsPing) Unmarshal(body []byte) *wsPing {
	_ = json.Unmarshal(body, p)

	return p
}
