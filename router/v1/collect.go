package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/entity"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

type CollectQuery struct {
	From     string `json:"from" form:"fsym" binding:"required,crypto"`
	To       string `json:"to" form:"tsym" binding:"required,common"`
	Interval uint   `json:"interval" form:"interval,default=60"`
}

func pulling(wc *dataproviders.Workers, query *CollectQuery) {
	wc.GetWorker(query.From, query.To).Interval = query.Interval
	if err := dataproviders.PullingData(wc, query.From, query.To); err != nil {
		handlers.SystemHandler(err)
	}
}

func AddWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res entity.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		worker := wc.AddWorker(query.From, query.To, wc.Pipe)
		if worker.IsAlive {
			res.UpdateAllFields(http.StatusOK, "Data for this pair is already being collected", nil)
		} else {
			go pulling(wc, &query)
			res.UpdateAllFields(http.StatusCreated, "Data collection started", nil)
		}
		return
	}
}

func RemoveWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res entity.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		worker := wc.GetWorker(query.From, query.To)
		if worker == nil || !worker.IsAlive {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", nil)
		} else {
			worker.Shutdown()
			res.UpdateAllFields(http.StatusOK, "Worker stopped successfully", nil)
		}
		return
	}
}

func WorkersStatus(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res entity.Result, err error) {
		res.UpdateAllFields(http.StatusOK, "Information about running workers", wc)
		return
	}
}

func UpdateWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res entity.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		worker := wc.GetWorker(query.From, query.To)
		if worker == nil || !worker.IsAlive {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", nil)
		} else {
			worker.Interval = query.Interval
			res.UpdateAllFields(http.StatusOK, "Worker updated successfully", worker)
		}
		return
	}
}
