package cryptocompare

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dataproviders"
)

const (
	apiURL = "https://min-api.cryptocompare.com"

	// Multiple Symbols Full Data - Get all the current trading info (price, vol, open, high, low etc) of any list of
	// cryptocurrencies in any other currency that you need. If the crypto does not trade directly into the toSymbol
	// requested, BTC will be used for conversion. This API also returns Display values for all the fields. If the
	// opposite pair trades we invert it (eg.: BTC-XMR)
	multipleSymbolsFullData = "/data/pricemultifull"
)

// CryptoCompareData structure for easily json serialization
type cryptoCompareData struct {
	Raw     map[string]map[string]*dataproviders.Response `json:"RAW"`
	Display map[string]map[string]*dataproviders.Display  `json:"DISPLAY"`
}

type cryptoCompare struct {
	apiKey string
}

// Init apiKey, apiURL, wsURL variables with environment values and return CryptoCompareData structure
func Init() (cc dataproviders.DataProvider, err error) {
	apiKey := config.GetEnv("CCDC_APIKEY")
	if apiKey == "" {
		return nil, errors.New("you should specify \"CCDC_APIKEY\" in you OS environment")
	}
	return &cryptoCompare{
		apiKey: apiKey,
	}, nil
}

// Get filled CryptoCompareData structure for the selected pair currencies over http/https
func (cc *cryptoCompare) Get(fSym string, tSym string) (ds *dataproviders.Data, err error) {
	var (
		apiUrl   *url.URL
		response *http.Response
		body     []byte
	)
	if apiUrl, err = cc.buildURL(fSym, tSym); err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: time.Duration(config.HttpClientTimeout) * time.Millisecond,
	}
	if response, err = client.Get(apiUrl.String()); err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, err
	}
	rawData := &cryptoCompareData{}
	if err = json.Unmarshal(body, rawData); err != nil {
		return nil, err
	}
	return convertToDomain(rawData), nil
}

func convertToDomain(d *cryptoCompareData) *dataproviders.Data {
	return (*dataproviders.Data)(d)
}

func (cc *cryptoCompare) buildURL(fSym string, tSym string) (u *url.URL, err error) {
	if u, err = url.Parse(apiURL + multipleSymbolsFullData); err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("fsyms", fSym)
	query.Set("tsyms", tSym)
	query.Set("api_key", cc.apiKey)
	u.RawQuery = query.Encode()
	return u, nil
}
