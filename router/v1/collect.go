package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/handlers"
	"net/http"
)

// CollectQuery structure for easily json serialization/validation/binding GET and POST query data
type CollectQuery struct {
	From     string `json:"fsym" form:"fsym" binding:"required,crypto"`
	To       string `json:"tsym" form:"tsym" binding:"required,common"`
	Interval int    `json:"interval" form:"interval,default=60"`
}

// AddWorker that will collect data for the selected currency pair to the management service
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
		worker = wc.AddWorker(query.From, query.To, query.Interval)
		worker.Work(wc.GetDataProvider())
		res.UpdateAllFields(http.StatusCreated, "Data collection started", worker)
		return
	}
}

// RemoveWorker from the management service and stop collecting data for the selected currencies pair
func RemoveWorker(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if worker := wc.GetWorker(query.From, query.To); worker == nil {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", nil)
			return
		}
		wc.RemoveWorker(query.From, query.To)
		res.UpdateAllFields(http.StatusOK, "Worker stopped successfully", nil)
		return
	}
}

// WorkersStatus return information about running workers
func WorkersStatus(wc *dataproviders.Workers) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		res.UpdateAllFields(http.StatusOK, "Information about running workers", nil)
		activeWorkers := wc.GetWorkers()
		if len(activeWorkers) == 0 {
			return
		}
		list := map[string]map[string]*dataproviders.Worker{}
		for worker := range activeWorkers {
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

// UpdateWorker update pulling data interval for the selected worker by the currencies pair
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
