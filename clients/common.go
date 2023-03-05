package clients

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
	Mktcap          float64 `json:"MKTCAP"`
	Lastupdate      int64   `json:"LASTUPDATE"`
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
	Lastupdate      string `json:"LASTUPDATE"`
	Supply          string `json:"SUPPLY"`
	Mktcap          string `json:"MKTCAP"`
}

// Data structure for easily json serialization
type Data struct {
	From    string
	To      string
	Raw     *Response `json:"RAW"`
	Display *Display  `json:"DISPLAY"`
}

// RestClient interface makes it possible to expand the list of data providers
type RestClient interface {
	Get(from string, to string) (*Data, error)
}

type WssClient interface {
	Subscribe(from string, to string) error
	Unsubscribe(from string, to string) error
}

// GetEmptyData returns empty Data
func GetEmptyData(from string, to string) *Data {
	return &Data{
		From:    from,
		To:      to,
		Raw:     &Response{},
		Display: &Display{},
	}
}
