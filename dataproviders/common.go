package dataproviders

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

type Data struct {
	Raw     map[string]map[string]*Response `json:"RAW"`
	Display map[string]map[string]*Display  `json:"DISPLAY"`
}

type DisplayData struct {
}

type DataPipe struct {
	From string
	To   string
	Data *Data
}

type DataProvider interface {
	GetSerializable() *Data
	GetData(from string, to string) (*Data, error)
}

func GetEmptyData(from string, to string) *Data {
	result := &Data{
		Raw:     map[string]map[string]*Response{},
		Display: map[string]map[string]*Display{},
	}
	result.Raw[from] = make(map[string]*Response)
	result.Raw[from][to] = &Response{}
	result.Display[from] = make(map[string]*Display)
	result.Display[from][to] = &Display{}
	return result
}
