package cryptocompare

import (
	"github.com/pkg/errors"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"net/url"
)

var apiKey string
var apiURL string

type CryptoCompare struct {
	Raw map[string]map[string]dataproviders.Response `json:"RAW"`
}

func (cc *CryptoCompare) GetSerializable() (dpStruct *dataproviders.Data) {
	return (*dataproviders.Data)(cc)
}

func (cc *CryptoCompare) GetApiURLToCompare(fSym string, tSym string) (u *url.URL, err error) {
	if apiURL == "" || apiKey == "" {
		return nil, errors.Wrap(err, "please, set OS environment \"CCDC_APIKEY\" and \"CCDC_APIURL\"")
	}
	if u, err = url.Parse(apiURL); err != nil {
		return nil, errors.Wrap(err, "can't parse raw url")
	}
	query := u.Query()
	query.Set("fsyms", fSym)
	query.Set("tsyms", tSym)
	query.Set("api_key", apiKey)
	u.RawQuery = query.Encode()
	return u, nil
}

func (cc *CryptoCompare) GetApiKey() string{
	return apiKey
}

func (cc *CryptoCompare) GetApiURL() string{
	return apiURL
}

func (cc *CryptoCompare) SetApiKey(key string) {
	apiKey = key
}

func (cc *CryptoCompare) SetApiURL(url string) {
	apiURL = url
}