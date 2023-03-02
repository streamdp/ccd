package router

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/dataproviders/cryptocompare"
	"github.com/streamdp/ccd/dbconnectors"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
	"github.com/streamdp/ccd/router/v1/validators"
	"github.com/streamdp/ccd/router/v1/ws"
)

// InitRouter basic work on setting up the application, declare endpoints, register our custom validation functions
func InitRouter(r *gin.Engine) (err error) {
	dp, err := cryptocompare.Init()
	if err != nil {
		return err
	}
	wc := dataproviders.NewWorkersControl(dp)
	db, err := dbconnectors.Connect()
	if err != nil {
		return err
	}
	dbconnectors.ServePipe(db, wc.GetPipe())

	r.LoadHTMLFiles("site/index.tmpl")
	r.Static("/css", "site/css")
	r.Static("/js", "site/js")
	r.GET("/", SendHTML)
	r.HEAD("/", SendOK)

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
		apiV1.GET("/ws", ws.HandleWs(wc, db))
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
