package domain

type Data struct {
	Id              int64   `db:"_id"             json:"id"`
	FromSymbol      string  `db:"fromSym"         json:"from_sym"`
	ToSymbol        string  `db:"toSym"           json:"to_sym"`
	Change24Hour    float64 `db:"change24hour"    json:"change_24_hour"`
	ChangePct24Hour float64 `db:"changepct24hour" json:"change_pct_24_hour"`
	Open24Hour      float64 `db:"open24hour"      json:"open_24_hour"`
	Volume24Hour    float64 `db:"volume24hour"    json:"volume_24_hour"`
	Low24Hour       float64 `db:"low24hour"       json:"low_24_hour"`
	High24Hour      float64 `db:"high24hour"      json:"high_24_hour"`
	Price           float64 `db:"price"           json:"price"`
	Supply          float64 `db:"supply"          json:"supply"`
	MktCap          float64 `db:"mktcap"          json:"mkt_cap"`
	LastUpdate      int64   `db:"lastupdate"      json:"last_update"`
	DisplayDataRaw  string  `db:"displaydataraw"  json:"display_data_raw"`
}

// Raw structure for easily json serialization
type Raw struct {
	FromSymbol      string  `json:"from_symbol"`
	ToSymbol        string  `json:"to_symbol"`
	Change24Hour    float64 `json:"change_24_hour"`
	ChangePct24Hour float64 `json:"changepct_24_hour"`
	Open24Hour      float64 `json:"open_24_hour"`
	Volume24Hour    float64 `json:"volume_24_hour"`
	Volume24HourTo  float64 `json:"volume_24_hour_to"`
	Low24Hour       float64 `json:"low_24_hour"`
	High24Hour      float64 `json:"high_24_hour"`
	Price           float64 `json:"price"`
	Supply          float64 `json:"supply"`
	MktCap          float64 `json:"mkt_cap"`
	LastUpdate      int64   `json:"last_update"`
}
