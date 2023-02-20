package cryptocompare

import (
	"encoding/json"
	"errors"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dataproviders"
	"io"
	"net/http"
	"net/url"
	"time"
)

// CryptoCompareData structure for easily json serialization
type data struct {
	Raw     map[string]map[string]*dataproviders.Response `json:"RAW"`
	Display map[string]map[string]*dataproviders.Display  `json:"DISPLAY"`
}

type cryptoCompare struct {
	apiKey string
	apiURL string
	wsURL  string
}

func convertToDomain(d *data) *dataproviders.Data {
	return (*dataproviders.Data)(d)
}

// Init apiKey, apiURL, wsURL variables with environment values and return CryptoCompareData structure
func Init() (cc dataproviders.DataProvider, err error) {
	apiKey := config.GetEnv("CCDC_APIKEY")
	apiURL := config.GetEnv("CCDC_APIURL")
	wsURL := config.GetEnv("CCDC_WSURL")
	if apiKey == "" || apiURL == "" {
		return nil, errors.New("you should specify \"CCDC_APIKEY\" and \"CCDC_APIURL\" in you OS environment")
	}
	return &cryptoCompare{
		apiKey: apiKey,
		apiURL: apiURL,
		wsURL:  wsURL,
	}, nil
}

// Get filled CryptoCompareData structure for the selected pair currencies over http/https
func (cc *cryptoCompare) Get(fSym string, tSym string) (ds *dataproviders.Data, err error) {
	var (
		apiUrl   *url.URL
		response *http.Response
		body     []byte
	)
	if apiUrl, err = cc.BuildURL(fSym, tSym); err != nil {
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
	rawData := &data{}
	if err = json.Unmarshal(body, rawData); err != nil {
		return nil, err
	}
	return convertToDomain(rawData), nil
}

// BuildURL for the selected pair currencies
func (cc *cryptoCompare) BuildURL(fSym string, tSym string) (u *url.URL, err error) {
	if u, err = url.Parse(cc.apiURL); err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("fsyms", fSym)
	query.Set("tsyms", tSym)
	query.Set("api_key", cc.apiKey)
	u.RawQuery = query.Encode()
	return u, nil
}
