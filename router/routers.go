package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/clients/cryptocompare"
	"github.com/streamdp/ccd/clients/huobi"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dbconnectors"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
	"github.com/streamdp/ccd/router/v1/validators"
	"github.com/streamdp/ccd/router/v1/ws"
)

// InitRouter basic work on setting up the application, declare endpoints, register our custom validation functions
func InitRouter(e *gin.Engine) (err error) {
	var (
		r clients.RestClient
		w clients.WssClient
	)
	switch config.DataProvider {
	case "huobi":
		r, err = huobi.Init()
	default:
		r, err = cryptocompare.Init()
	}
	if err != nil {
		return err
	}
	wc := clients.NewWorkersControl(r)
	db, err := dbconnectors.Connect()
	if err != nil {
		return err
	}
	dbconnectors.ServePipe(db, wc.GetPipe())

	if config.DataProvider == "huobi" {
		w = huobi.InitWs(wc.GetPipe())
	}

	e.LoadHTMLFiles("site/index.tmpl")
	e.Static("/css", "site/css")
	e.Static("/js", "site/js")
	e.GET("/", SendHTML)
	e.HEAD("/", SendOK)

	apiV1 := e.Group("/v1")
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
		apiV1.GET("/ws", ws.HandleWs(wc, db))
		if w != nil {
			apiV1.POST("/ws/subscribe", handlers.GinHandler(v1.Subscribe(w)))
			apiV1.GET("/ws/subscribe", handlers.GinHandler(v1.Subscribe(w)))
			apiV1.POST("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(w)))
			apiV1.GET("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(w)))
		}
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
