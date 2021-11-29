package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

type CollectQuery struct {
	From     string        `json:"from" form:"fsym" binding:"required,crypto"`
	To       string        `json:"to" form:"tsym" binding:"required,common"`
	Interval int           `json:"interval" form:"interval,default=60"`
}

func AddWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		var worker *dataproviders.Worker
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if worker = wc.GetWorker(query.From, query.To); worker != nil {
			res.UpdateAllFields(http.StatusOK, "Data for this pair is already being collected", worker)
			return
		}
		worker = wc.Add(wc.NewWorker(query.From, query.To))
		go worker.Work(wc.GetDataProvider())
		res.UpdateAllFields(http.StatusCreated, "Data collection started", worker)
		return
	}
}

func RemoveWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		var worker *dataproviders.Worker
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if worker = wc.GetWorker(query.From, query.To); worker == nil {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", nil)
			return
		}
		wc.RemoveWorker(query.From, query.To)
		res.UpdateAllFields(http.StatusOK, "Worker stopped successfully", nil)
		return
	}
}

func WorkersStatus(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		res.UpdateAllFields(http.StatusOK, "Information about running workers", nil)
		if len(*wc.GetWorkers()) == 0 {
			return
		}
		list := map[string]map[string]*dataproviders.Worker{}
		for worker := range *wc.GetWorkers() {
			if list[worker.From] != nil {
				list[worker.From][worker.To] = worker
				continue
			}
			list[worker.From] = make(map[string]*dataproviders.Worker)
			list[worker.From][worker.To] = worker
		}
		res.UpdateDataField(list)
		return
	}
}

func UpdateWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		var worker *dataproviders.Worker
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if worker = wc.GetWorker(query.From, query.To); worker == nil {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", worker)
			return
		}
		worker.Interval = query.Interval
		res.UpdateAllFields(http.StatusOK, "Worker updated successfully", worker)
		return
	}
}
