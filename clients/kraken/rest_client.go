package kraken

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
)

const (
	apiUrl = "https://api.kraken.com"

	// Get Ticker Information https://docs.kraken.com/api/docs/rest-api/get-ticker-information
	// Get ticker information for all or requested markets. To clarify usage, note that:
	// - Today's prices start at midnight UTC
	// - Leaving the pair parameter blank will return tickers for all tradeable assets on Kraken
	// Request Parameters "pair" (Asset pair to get data for (optional, default: all tradeable exchange pairs), e.g.
	// XBTUSD, WBTCUSD. Refer to /0/public/Assets)
	tickerInformation = "/0/public/Ticker"

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
	if len(rawData.Error) != 0 {
		return nil, fmt.Errorf("server error: %v", rawData.Error)
	}

	return convertRestDataToDomain(fSym, tSym, rawData, time.Now().UTC().UnixMilli()), nil
}

func (r *rest) limitRate() {
	<-r.limiter.C
	r.limiter.Reset(rateLimit)
}

func (r *rest) buildURL(fSym string, tSym string) (*url.URL, error) {
	u, err := url.Parse(apiUrl + tickerInformation)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}
	query := u.Query()
	query.Set("pair", strings.ToLower(fSym+tSym))
	u.RawQuery = query.Encode()

	return u, nil
}

func convertRestDataToDomain(from, to string, d *restData, lastUpdate int64) *domain.Data {
	if d == nil || len(d.Result) == 0 {
		return nil
	}

	var tick restTickerInfo
	for _, v := range d.Result {
		tick = v
	}

	open24Hour, _ := strconv.ParseFloat(tick.O, 64)
	volume24Hour, _ := strconv.ParseFloat(tick.V[1], 64)
	low24Hour, _ := strconv.ParseFloat(tick.L[1], 64)
	high24Hour, _ := strconv.ParseFloat(tick.H[1], 64)
	price, _ := strconv.ParseFloat(tick.P[0], 64)

	b, _ := json.Marshal(&domain.Raw{
		FromSymbol:   from,
		ToSymbol:     to,
		Open24Hour:   open24Hour,
		Volume24Hour: volume24Hour,
		Low24Hour:    low24Hour,
		High24Hour:   high24Hour,
		Price:        price,
		Supply:       float64(tick.T[1]),
		LastUpdate:   lastUpdate,
	})

	return &domain.Data{
		FromSymbol:     from,
		ToSymbol:       to,
		Open24Hour:     open24Hour,
		Volume24Hour:   volume24Hour,
		Low24Hour:      low24Hour,
		High24Hour:     high24Hour,
		Price:          price,
		Supply:         float64(tick.T[1]),
		LastUpdate:     lastUpdate,
		DisplayDataRaw: string(b),
	}
}
