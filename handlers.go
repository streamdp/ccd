package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/utility"
	"net/http"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func collectAdd(dp dataproviders.DataProvider, wc *dataproviders.Workers, pipe chan *dataproviders.DataPipe) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := utility.CollectQuery{}
		if err := c.BindJSON(&query); err != nil {
			utility.HandleError(errors.Wrapf(err, "can't bind json data"))
			return
		}
		if !utility.ValidateCollectQuery(query) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Missing values (look, interval should be >=60 or you set incorrect crypto/common values)",
			})
			return
		}
		worker := wc.AddWorker(query.From, query.To, pipe)
		if worker.IsAlive() {
			c.JSON(http.StatusOK, gin.H{
				"message": "Pair [" + query.From + "]-[" + query.To + "] already collecting",
			})
		} else {
			go func(wc *dataproviders.Worker) {
				if err := dataproviders.PullingData(dp, query.From, query.To, query.Interval, wc); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": "Can't start worker for the [" + query.From + "]-[" + query.To + "] pair",
					})
					return
				}
			}(worker)
			c.JSON(http.StatusOK, gin.H{
				"message": "Start pulling data for the [" + query.From + "]-[" + query.To + "] pair",
			})
		}
	}
}

func collectRemove(wc *dataproviders.Workers) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := utility.CollectQuery{}
		if err := c.BindJSON(&query); err != nil {
			utility.HandleError(errors.Wrapf(err, "can't bind json data"))
			return
		}
		if !utility.ValidateCollectQuery(query) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Missing values (look, interval should be >=60 or you set incorrect crypto/common values)",
			})
			return
		}
		worker := wc.GetWorker(query.From, query.To)
		if worker == nil || !worker.IsAlive() {
			c.JSON(http.StatusOK, gin.H{
				"message": "Pair [" + query.From + "]-[" + query.To + "] not collecting now",
			})
		} else {
			worker.Shutdown()
			c.JSON(http.StatusOK, gin.H{
				"message": "Worker [" + query.From + "]-[" + query.To + "] stopped successfully",
			})
		}
	}
}
