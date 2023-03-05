package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/dbconnectors"
	"github.com/streamdp/ccd/handlers"
)

// PriceQuery structure for easily json serialization/validation/binding GET and POST query data
type PriceQuery struct {
	From string `json:"fsym" form:"fsym" binding:"required,crypto"`
	To   string `json:"tsym" form:"tsym" binding:"required,common"`
}

// GetLastPrice return up-to-date data for the selected currencies pair
func GetLastPrice(p *clients.RestPuller, db dbconnectors.DbReadWrite, query *PriceQuery) (data *clients.Data, err error) {
	data, err = (*p.Client()).Get(query.From, query.To)
	if err != nil {
		if data, err = db.GetLast(query.From, query.To); err != nil {
			return
		}
	}
	p.DataPipe() <- data
	return
}

// GetPrice return up-to-date or most recent data for the selected currencies pair
func GetPrice(wc *clients.RestPuller, db dbconnectors.DbReadWrite) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := PriceQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		data, err := GetLastPrice(wc, db, &query)
		if err != nil {
			return
		}
		res.UpdateAllFields(http.StatusOK, "Most recent price, updated at "+data.Display.Lastupdate, data)
		return
	}
}
