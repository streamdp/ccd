package v1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/server/handlers"
)

// PriceQuery structure for easily json serialization/validation/binding GET and POST query data
type PriceQuery struct {
	From string `binding:"required,symbols" form:"fsym" json:"fsym"`
	To   string `binding:"required,symbols" form:"tsym" json:"tsym"`
}

func (p *PriceQuery) ToUpper() *PriceQuery {
	p.From = strings.ToUpper(p.From)
	p.To = strings.ToUpper(p.To)

	return p
}

var errGetPrice = errors.New("failed to get price")

// LastPrice return up-to-date data for the selected currencies pair
func LastPrice(r clients.RestClient, db db.Database, from, to string) (*domain.Data, error) {
	data, err := r.Get(from, to)
	if err != nil {
		if data, err = db.GetLast(from, to); err != nil {
			return nil, errGetPrice
		}

		return data, nil
	}

	db.DataPipe() <- data

	return data, nil
}

// Price return up-to-date or most recent data for the selected currencies pair
func Price(rc clients.RestClient, db db.Database) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := PriceQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.ToUpper()

		p, err := LastPrice(rc, db, q.From, q.To)
		if err != nil {
			return &domain.Result{}, fmt.Errorf("failed to get price: %w", err)
		}

		return domain.NewResult(
			http.StatusOK,
			fmt.Sprintf("Most recent price, updated at %d", p.LastUpdate),
			p,
		), nil
	}
}
