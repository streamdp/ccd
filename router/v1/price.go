package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/dbconnectors"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

type PriceQuery struct {
	From string `json:"from" form:"fsym" binding:"required,crypto"`
	To   string `json:"to" form:"tsym" binding:"required,common"`
}

func GetLastPrice(wc *dataproviders.Workers, db *dbconnectors.Db) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := PriceQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		data, err := wc.Dp.GetData(query.From, query.To)
		if err != nil {
			data, err = db.GetLast(query.From, query.To)
			if err != nil {
				return
			}
		}
		wc.Pipe <- &dataproviders.DataPipe{
			From: query.From,
			To:   query.To,
			Data: data,
		}
		res.UpdateAllFields(http.StatusOK, "Most recent price, updated at "+
			data.Display[query.From][query.To].Lastupdate, data)
		return
	}
}
