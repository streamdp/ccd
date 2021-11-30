package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/dataproviders/cryptocompare"
	//	"github.com/streamdp/ccdatacollector/dataproviders/ws"
	"github.com/streamdp/ccdatacollector/dbconnectors"
	"github.com/streamdp/ccdatacollector/handlers"
	v1 "github.com/streamdp/ccdatacollector/router/v1"
	"github.com/streamdp/ccdatacollector/router/v1/validators"
	ws1 "github.com/streamdp/ccdatacollector/router/v1/ws"
)

func InitRouter(r *gin.Engine) (err error) {
	dp, err := cryptocompare.Init()
	if err != nil {
		return err
	}
	wc := dataproviders.NewWorkersControl(dp)
	go wc.Run()
	db, err := dbconnectors.Connect()
	if err != nil {
		return err
	}
	go db.ServePipe(wc.GetPipe())
	hub := ws1.NewHub()
	go hub.Run(wc, db)
	//- only for testing (it printing data to the log)-//
	//	hub2 := ws.NewHub()
	//	go hub2.Run()
	//	wsClient := ws.ServeWs(hub2, dp)
	//	wsClient.Subscribe([]string{"5~CCCAGG~BTC~USD"})
	//-------------------------------------------------//
	r.LoadHTMLFiles("site/index.tmpl")
	r.Static("/css", "site/css")
	r.Static("/js", "site/js")
	r.GET("/", IndexHTML)
	apiV1 := r.Group("/v1")
	{
		apiV1.GET("/service/ping", handlers.GinHandler(v1.Ping))
		apiV1.POST("/collect/add", handlers.GinHandler(v1.AddWorker(wc)))
		apiV1.GET("/collect/add", handlers.GinHandler(v1.AddWorker(wc)))
		apiV1.POST("/collect/remove", handlers.GinHandler(v1.RemoveWorker(wc)))
		apiV1.GET("/collect/remove", handlers.GinHandler(v1.RemoveWorker(wc)))
		apiV1.GET("/collect/status", handlers.GinHandler(v1.WorkersStatus(wc)))
		apiV1.POST("/collect/update", handlers.GinHandler(v1.UpdateWorker(wc)))
		apiV1.GET("/collect/update", handlers.GinHandler(v1.UpdateWorker(wc)))
		apiV1.POST("/price", handlers.GinHandler(v1.GetPrice(wc, db)))
		apiV1.GET("/price", handlers.GinHandler(v1.GetPrice(wc, db)))
		apiV1.GET("/ws", ws1.ServeWs(hub))
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
