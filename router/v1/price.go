package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/dbconnectors"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

type PriceQuery struct {
	From string `json:"fsym" form:"fsym" binding:"required,crypto"`
	To   string `json:"tsym" form:"tsym" binding:"required,common"`
}

func GetLastPrice(wc *dataproviders.Workers, db *dbconnectors.Db, query *PriceQuery) (data *dataproviders.Data, err error) {
	data, err = (*wc.GetDataProvider()).GetData(query.From, query.To)
	if err != nil {
		data, err = db.GetLast(query.From, query.To)
		if err != nil {
			return
		}
	}
	if data == (dataproviders.GetEmptyData(query.From, query.To)) {
		return
	}
	wc.GetPipe() <- &dataproviders.DataPipe{
		From: query.From,
		To:   query.To,
		Data: data,
	}
	return
}

func GetPrice(wc *dataproviders.Workers, db *dbconnectors.Db) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := PriceQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		data, err := GetLastPrice(wc, db, &query)
		res.UpdateAllFields(http.StatusOK, "Most recent price, updated at "+
			data.Display[query.From][query.To].Lastupdate, data)
		return
	}
}
