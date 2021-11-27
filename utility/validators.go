package utility

var (
	cryptoCurrencies = []string{"BTC", "XRP", "ETH", "BCH", "EOS", "LTC", "XMR", "DASH"}
	commonCurrencies = []string{"USD", "EUR", "GBP", "JPY", "RUR"}
)

type CollectQuery struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int    `json:"interval"`
}

func isCrypto(s string) bool {
	for _, val := range cryptoCurrencies {
		if val == s {
			return true
		}
	}
	return false
}

func isCommon(s string) bool {
	for _, val := range commonCurrencies {
		if val == s {
			return true
		}
	}
	return false
}

func ValidateCollectQuery(query CollectQuery) (valid bool) {
	return isCrypto(query.From) && isCommon(query.To) && query.Interval >= 60
}
