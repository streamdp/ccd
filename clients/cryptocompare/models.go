package cryptocompare

import "github.com/streamdp/ccd/clients"

type cryptoCompareData struct {
	Raw     map[string]map[string]*clients.Response `json:"RAW"`
	Display map[string]map[string]*clients.Display  `json:"DISPLAY"`
}

type cryptoCompareWsData struct {
	Type                string  `json:"TYPE"`
	FromSymbol          string  `json:"FROMSYMBOL"`
	ToSymbol            string  `json:"TOSYMBOL"`
	Price               float64 `json:"PRICE"`
	LastUpdate          int64   `json:"LASTUPDATE"`
	Volume24Hour        float64 `json:"VOLUME24HOUR"`
	Volume24HourTo      float64 `json:"VOLUME24HOURTO"`
	Open24Hour          float64 `json:"OPEN24HOUR"`
	High24Hour          float64 `json:"HIGH24HOUR"`
	Low24Hour           float64 `json:"LOW24HOUR"`
	CurrentSupply       float64 `json:"CURRENTSUPPLY"`
	CurrentSupplyMktCap float64 `json:"CURRENTSUPPLYMKTCAP"`
}
