package huobi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
)

const (
	apiUrl = "https://api.huobi.pro"

	// Get Latest Aggregated Ticker https://huobiapi.github.io/docs/spot/v1/en/#get-latest-aggregated-ticker
	// This endpoint retrieves the latest ticker with some important 24h aggregated market data.
	// Request Parameters "symbol" (all supported trading symbol, e.g. btcusdt, bccbtc. Refer to /v1/common/symbols)
	latestAggregatedTicker = "/market/detail/merged"

	rateLimit = 500 * time.Millisecond // make max two api calls per second
)

type rest struct {
	client  *http.Client
	limiter *time.Timer
}

var errWrongStatusCode = errors.New("wrong response status code")

func Init(cfg *config.App) (*rest, error) {
	return &rest{
		client: &http.Client{
			Timeout: cfg.Http.ClientTimeout(),
		},
		limiter: time.NewTimer(rateLimit),
	}, nil
}

func (r *rest) Get(fSym string, tSym string) (*domain.Data, error) {
	r.limitRate()

	var (
		response *http.Response
		body     []byte
	)
	u, err := r.buildURL(fSym, tSym)
	if err != nil {
		return nil, fmt.Errorf("failed to build url: %w", err)
	}

	if response, err = r.client.Get(u.String()); err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, errWrongStatusCode
	}
	rawData := &restData{}
	if err = json.Unmarshal(body, rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if rawData.Status == "error" {
		return nil, fmt.Errorf("server error: %v", rawData.ErrMsg)
	}

	return convertRestDataToDomain(fSym, tSym, rawData), nil
}

func (r *rest) limitRate() {
	<-r.limiter.C
	r.limiter.Reset(rateLimit)
}

func (r *rest) buildURL(fSym string, tSym string) (*url.URL, error) {
	u, err := url.Parse(apiUrl + latestAggregatedTicker)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	if strings.ToLower(tSym) == "usd" {
		tSym = "usdt"
	}
	query := u.Query()
	query.Set("symbol", strings.ToLower(fSym+tSym))
	u.RawQuery = query.Encode()

	return u, nil
}

func convertRestDataToDomain(from, to string, d *restData) *domain.Data {
	if d == nil {
		return nil
	}
	var price float64
	if len(d.Tick.Bid) > 0 {
		price = d.Tick.Bid[0]
	}
	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     d.Tick.Open,
		Volume24Hour:   d.Tick.Amount,
		Volume24HourTo: d.Tick.Vol,
		High24Hour:     d.Tick.High,
		Price:          price,
		LastUpdate:     d.Ts,
		Supply:         float64(d.Tick.Count),
	})

	return &domain.Data{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     d.Tick.Open,
		Volume24Hour:   d.Tick.Amount,
		Low24Hour:      d.Tick.Low,
		High24Hour:     d.Tick.High,
		Price:          price,
		Supply:         float64(d.Tick.Count),
		LastUpdate:     d.Ts,
		DisplayDataRaw: string(b),
	}
}
