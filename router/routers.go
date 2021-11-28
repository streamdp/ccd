package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/dataproviders/cryptocompare"
	"github.com/streamdp/ccdatacollector/dbconnectors"
	"github.com/streamdp/ccdatacollector/handlers"
	v1 "github.com/streamdp/ccdatacollector/router/v1"
	"github.com/streamdp/ccdatacollector/router/validators"
)

func InitRouter(r *gin.Engine) (err error) {
	dp, err := cryptocompare.Init()
	if err != nil {
		return err
	}
	wc := dataproviders.CreateWorkersControl(dp)
	db, err := dbconnectors.Connect()
	if err != nil {
		return err
	}
	go db.ServePipe(wc.Pipe)
	apiGroupV1 := r.Group("/v1")
	{
		apiGroupV1.GET("/service/ping", handlers.GinHandler(v1.Ping()))
		apiGroupV1.POST("/collect/add", handlers.GinHandler(v1.AddWorker(wc)))
		apiGroupV1.GET("/collect/add", handlers.GinHandler(v1.AddWorker(wc)))
		apiGroupV1.POST("/collect/remove", handlers.GinHandler(v1.RemoveWorker(wc)))
		apiGroupV1.GET("/collect/remove", handlers.GinHandler(v1.RemoveWorker(wc)))
		apiGroupV1.GET("/collect/status", handlers.GinHandler(v1.WorkersStatus(wc)))
		apiGroupV1.POST("/collect/update", handlers.GinHandler(v1.UpdateWorker(wc)))
		apiGroupV1.GET("/collect/update", handlers.GinHandler(v1.UpdateWorker(wc)))
		apiGroupV1.POST("/price", handlers.GinHandler(v1.GetLastPrice(wc, db)))
		apiGroupV1.GET("/price", handlers.GinHandler(v1.GetLastPrice(wc, db)))
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err = v.RegisterValidation("crypto", validators.Crypto); err != nil {
			return err
		}
		if err = v.RegisterValidation("common", validators.Common); err != nil {
			return err
		}
	}
	return nil
}
