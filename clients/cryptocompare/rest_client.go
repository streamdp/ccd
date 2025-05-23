package cryptocompare

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
)

const (
	apiUrl = "https://min-api.cryptocompare.com"

	// Multiple Symbols Full Data - Get all the current trading info (price, vol, open, high, low etc) of any list of
	// cryptocurrencies in any other currency that you need. If the crypto does not trade directly into the toSymbol
	// requested, BTC will be used for conversion. This API also returns Display values for all the fields. If the
	// opposite pair trades we invert it (eg.: BTC-XMR)
	multipleSymbolsFullData = "/data/pricemultifull"
)

type rest struct {
	apiKey string
	client *http.Client
}

var (
	errApiKeyNotDefined = errors.New("you should specify \"CCDC_APIKEY\" in you OS environment")
	errWrongStatusCode  = errors.New("wrong status code")
)

// Init apiKey, apiUrl, wsURL variables with environment values and return CryptoCompareData structure
func Init(cfg *config.App) (*rest, error) {
	if cfg.ApiKey == "" {
		return nil, errApiKeyNotDefined
	}

	return &rest{
		apiKey: cfg.ApiKey,
		client: &http.Client{
			Timeout: cfg.Http.ClientTimeout(),
		},
	}, nil
}

// Get filled CryptoCompareData structure for the selected pair currencies over http/https
func (r *rest) Get(fSym string, tSym string) (*domain.Data, error) {
	var (
		response *http.Response
		body     []byte
	)
	u, err := r.buildURL(fSym, tSym)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %w", err)
	}
	response, err = r.client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	body, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, errWrongStatusCode
	}
	rawData := &restData{}
	if err = json.Unmarshal(body, rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return convertToDomain(fSym, tSym, rawData), nil
}

func (r *rest) buildURL(fSym string, tSym string) (*url.URL, error) {
	u, err := url.Parse(apiUrl + multipleSymbolsFullData)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %w", err)
	}
	query := u.Query()
	query.Set("fsyms", fSym)
	query.Set("tsyms", tSym)
	query.Set("api_key", r.apiKey)
	u.RawQuery = query.Encode()

	return u, nil
}

func convertToDomain(from, to string, d *restData) *domain.Data {
	if d == nil || d.Raw == nil || d.Raw[from] == nil {
		return nil
	}
	r := d.Raw[from][to]
	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     r.Open24Hour,
		Volume24Hour:   r.Volume24Hour,
		Volume24HourTo: r.Volume24Hourto,
		High24Hour:     r.High24Hour,
		Price:          r.Price,
		LastUpdate:     r.LastUpdate,
		Supply:         r.Supply,
		MktCap:         r.MktCap,
	})

	return &domain.Data{
		FromSymbol:      from,
		ToSymbol:        to,
		Change24Hour:    r.Change24Hour,
		ChangePct24Hour: r.Changepct24Hour,
		Open24Hour:      r.Open24Hour,
		Volume24Hour:    r.Volume24Hour,
		Low24Hour:       r.Low24Hour,
		High24Hour:      r.High24Hour,
		Price:           r.Price,
		Supply:          r.Supply,
		MktCap:          r.MktCap,
		LastUpdate:      r.LastUpdate,
		DisplayDataRaw:  string(b),
	}
}
