package cryptocompare

import (
	"encoding/json"
	"errors"
	"github.com/streamdp/ccdatacollector/config"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/utility"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var apiKey string
var apiURL string

type Data struct {
	Raw     map[string]map[string]*dataproviders.Response `json:"RAW"`
	Display map[string]map[string]*dataproviders.Display  `json:"DISPLAY"`
}

func (cc *Data) GetSerializable() (dpStruct *dataproviders.Data) {
	return (*dataproviders.Data)(cc)
}

func Init() (cc *Data, err error) {
	cc.SetApiKey(utility.GetEnv("CCDC_APIKEY"))
	cc.SetApiURL(utility.GetEnv("CCDC_APIURL"))
	if cc.GetApiURL() == "" || cc.GetApiKey() == "" {
		return nil, errors.New("you should specify \"CCDC_APIKEY\" and \"CCDC_APIURL\" in you OS environment")
	}
	return cc, nil
}

func (cc *Data) GetData(fSym string, tSym string) (ds *dataproviders.Data, err error) {
	var apiURL *url.URL
	var response *http.Response
	var data []byte
	if apiURL, err = cc.BuildURL(fSym, tSym); err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: config.HttpClientTimeout * time.Millisecond,
	}
	if response, err = client.Get(apiURL.String()); err != nil {
		return nil, err
	}
	if data, err = ioutil.ReadAll(response.Body); err != nil {
		return nil, err
	}
	if response.StatusCode == 200 {
		if err = json.Unmarshal(data, &cc); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return cc.GetSerializable(), nil
}

func (cc *Data) BuildURL(fSym string, tSym string) (u *url.URL, err error) {
	if apiURL == "" || apiKey == "" {
		return nil, errors.New("please, set OS environment \"CCDC_APIKEY\" and \"CCDC_APIURL\"")
	}
	if u, err = url.Parse(apiURL); err != nil {
		return nil, err
	}
	query := u.Query()
	query.Set("fsyms", fSym)
	query.Set("tsyms", tSym)
	query.Set("api_key", apiKey)
	u.RawQuery = query.Encode()
	return u, nil
}

func (cc *Data) GetApiKey() string {
	return apiKey
}

func (cc *Data) GetApiURL() string {
	return apiURL
}

func (cc *Data) SetApiKey(key string) {
	apiKey = key
}

func (cc *Data) SetApiURL(url string) {
	apiURL = url
}
