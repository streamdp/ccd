package huobi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dataproviders"
)

const (
	apiURL = "https://api.huobi.pro"

	// Get Latest Aggregated Ticker https://huobiapi.github.io/docs/spot/v1/en/#get-latest-aggregated-ticker
	// This endpoint retrieves the latest ticker with some important 24h aggregated market data.
	// Request Parameters "symbol" (all supported trading symbol, e.g. btcusdt, bccbtc. Refer to /v1/common/symbols)
	latestAggregatedTicker = "/market/detail/merged"
)

type huobiData struct {
	Ch      string `json:"ch"`
	Status  string `json:"status"`
	ErrCode string `json:"err-code"`
	ErrMsg  string `json:"err-msg"`
	Ts      int64  `json:"ts"`
	Tick    struct {
		Id      int64     `json:"id"`
		Version int64     `json:"version"`
		Open    float64   `json:"open"`
		Close   float64   `json:"close"`
		Low     float64   `json:"low"`
		High    float64   `json:"high"`
		Amount  float64   `json:"amount"`
		Vol     float64   `json:"vol"`
		Count   int       `json:"count"`
		Bid     []float64 `json:"bid"`
		Ask     []float64 `json:"ask"`
	} `json:"tick"`
}

type huobi struct{}

func Init() (dataproviders.DataProvider, error) {
	return &huobi{}, nil
}

func (cc *huobi) Get(fSym string, tSym string) (ds *dataproviders.Data, err error) {
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
	rawData := &huobiData{}
	if err = json.Unmarshal(body, rawData); err != nil {
		return nil, err
	}
	if rawData.Status == "error" {
		return nil, errors.New(rawData.ErrMsg)
	}
	return convertToDomain(fSym, tSym, rawData), nil
}

func convertToDomain(from, to string, d *huobiData) *dataproviders.Data {
	if d == nil {
		return nil
	}
	var price float64
	if len(d.Tick.Bid) > 0 {
		price = d.Tick.Bid[0]
	}
	return &dataproviders.Data{
		Raw: map[string]map[string]*dataproviders.Response{
			strings.ToUpper(from): {
				strings.ToUpper(to): {
					Open24Hour:     d.Tick.Open,
					Volume24Hour:   d.Tick.Amount,
					Volume24Hourto: d.Tick.Vol,
					Low24Hour:      d.Tick.Low,
					High24Hour:     d.Tick.High,
					Price:          price,
					Supply:         float64(d.Tick.Count),
					Lastupdate:     d.Ts,
				},
			},
		},
		Display: map[string]map[string]*dataproviders.Display{
			strings.ToUpper(from): {
				strings.ToUpper(to): {
					Open24Hour:     strconv.FormatFloat(d.Tick.Open, 'f', -1, 64),
					Volume24Hour:   strconv.FormatFloat(d.Tick.Amount, 'f', -1, 64),
					Volume24Hourto: strconv.FormatFloat(d.Tick.Vol, 'f', -1, 64),
					High24Hour:     strconv.FormatFloat(d.Tick.High, 'f', -1, 64),
					Price:          strconv.FormatFloat(price, 'f', -1, 64),
					FromSymbol:     strings.ToUpper(from),
					ToSymbol:       strings.ToUpper(to),
					Lastupdate:     strconv.FormatInt(d.Ts, 10),
					Supply:         strconv.Itoa(d.Tick.Count),
				},
			},
		},
	}
}

func (cc *huobi) buildURL(fSym string, tSym string) (u *url.URL, err error) {
	if u, err = url.Parse(apiURL + latestAggregatedTicker); err != nil {
		return nil, err
	}
	if strings.ToLower(tSym) == "usd" {
		tSym = "usdt"
	}
	query := u.Query()
	query.Set("symbol", strings.ToLower(fSym+tSym))
	u.RawQuery = query.Encode()
	return u, nil
}
