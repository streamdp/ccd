package clients

import "strings"

// Response structure for easily json serialization
type Response struct {
	Change24Hour    float64 `json:"CHANGE24HOUR"`
	Changepct24Hour float64 `json:"CHANGEPCT24HOUR"`
	Open24Hour      float64 `json:"OPEN24HOUR"`
	Volume24Hour    float64 `json:"VOLUME24HOUR"`
	Volume24Hourto  float64 `json:"VOLUME24HOURTO"`
	Low24Hour       float64 `json:"LOW24HOUR"`
	High24Hour      float64 `json:"HIGH24HOUR"`
	Price           float64 `json:"PRICE"`
	Supply          float64 `json:"SUPPLY"`
	MktCap          float64 `json:"MKTCAP"`
	LastUpdate      int64   `json:"LASTUPDATE"`
}

// Display structure for easily json serialization
type Display struct {
	Change24Hour    string `json:"CHANGE24HOUR"`
	Changepct24Hour string `json:"CHANGEPCT24HOUR"`
	Open24Hour      string `json:"OPEN24HOUR"`
	Volume24Hour    string `json:"VOLUME24HOUR"`
	Volume24Hourto  string `json:"VOLUME24HOURTO"`
	High24Hour      string `json:"HIGH24HOUR"`
	Price           string `json:"PRICE"`
	FromSymbol      string `json:"FROMSYMBOL"`
	ToSymbol        string `json:"TOSYMBOL"`
	LastUpdate      string `json:"LASTUPDATE"`
	Supply          string `json:"SUPPLY"`
	MktCap          string `json:"MKTCAP"`
}

// Data structure for easily json serialization
type Data struct {
	From    string
	To      string
	Raw     *Response `json:"RAW"`
	Display *Display  `json:"DISPLAY"`
}

// EmptyData returns empty Data
func EmptyData(from string, to string) *Data {
	return &Data{
		From:    from,
		To:      to,
		Raw:     &Response{},
		Display: &Display{},
	}
}

type Subscribe struct {
	From string `json:"from"`
	To   string `json:"to"`
	id   int64
}
type Subscribes map[string]*Subscribe

func NewSubscribe(from, to string, id int64) *Subscribe {
	return &Subscribe{
		From: strings.ToUpper(from),
		To:   strings.ToUpper(to),
		id:   id,
	}
}

func (s *Subscribe) Id() int64 {
	return s.id
}
