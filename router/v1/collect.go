package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/handlers"
)

// CollectQuery structure for easily json serialization/validation/binding GET and POST query data
type CollectQuery struct {
	From     string `json:"fsym" form:"fsym" binding:"required,crypto"`
	To       string `json:"tsym" form:"tsym" binding:"required,common"`
	Interval int    `json:"interval" form:"interval,default=60"`
}

// AddWorker that will collect data for the selected currency pair to the management service
func AddWorker(p *clients.RestPuller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		var w *clients.Worker
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if w = p.Worker(query.From, query.To); w != nil {
			res.UpdateAllFields(http.StatusOK, "Data for this pair is already being collected", w)
			return
		}
		w = p.AddWorker(query.From, query.To, query.Interval)
		res.UpdateAllFields(http.StatusCreated, "Data collection started", w)
		return
	}
}

// RemoveWorker from the management service and stop collecting data for the selected currencies pair
func RemoveWorker(p *clients.RestPuller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if p.Worker(query.From, query.To) == nil {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", nil)
			return
		}
		p.RemoveWorker(query.From, query.To)
		res.UpdateAllFields(http.StatusOK, "Worker stopped successfully", nil)
		return
	}
}

// WorkersStatus return information about running workers
func WorkersStatus(p *clients.RestPuller, w clients.WssClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		res.UpdateAllFields(http.StatusOK, "Information about running workers", nil)
		activeWorkers := p.ListWorkers()
		for k, v := range w.ListSubscribes() {
			activeWorkers[&clients.Worker{
				From: k.From,
				To:   k.To,
			}] = v
		}
		if len(activeWorkers) == 0 {
			return
		}
		list := map[string]map[string]*clients.Worker{}
		for worker := range activeWorkers {
			if list[worker.From] != nil {
				list[worker.From][worker.To] = worker
				continue
			}
			list[worker.From] = make(map[string]*clients.Worker)
			list[worker.From][worker.To] = worker
		}
		res.UpdateDataField(list)
		return
	}
}

// UpdateWorker update pulling data interval for the selected worker by the currencies pair
func UpdateWorker(p *clients.RestPuller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		var w *clients.Worker
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if w = p.Worker(query.From, query.To); w == nil {
			res.UpdateAllFields(http.StatusOK, "No data is collected for this pair", w)
			return
		}
		w.Interval = query.Interval
		res.UpdateAllFields(http.StatusOK, "Worker updated successfully", w)
		return
	}
}

func Subscribe(w clients.WssClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if err = w.Subscribe(query.From, query.To); err != nil {
			res.UpdateAllFields(http.StatusOK, "subscribe error:", err)
			return
		}
		res.UpdateAllFields(http.StatusCreated, "Subscribed successfully, data collection started", []string{query.From, query.To})
		return
	}
}

func Unsubscribe(w clients.WssClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (res handlers.Result, err error) {
		query := CollectQuery{}
		if err = c.Bind(&query); err != nil {
			return
		}
		if err = w.Unsubscribe(query.From, query.To); err != nil {
			res.UpdateAllFields(http.StatusOK, "Unsubscribe error:", err)
			return
		}
		res.UpdateAllFields(http.StatusOK, "Unsubscribed successfully, data collection stopped ", []string{query.From, query.To})
		return
	}
}
