package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccd/session"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/clients/cryptocompare"
	"github.com/streamdp/ccd/clients/huobi"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/handlers"
	v1 "github.com/streamdp/ccd/router/v1"
	"github.com/streamdp/ccd/router/v1/validators"
	"github.com/streamdp/ccd/router/v1/ws"
)

// InitRouter basic work on setting up the application, declare endpoints, register our custom validation functions
func InitRouter(e *gin.Engine, s *session.KeysStore) (err error) {
	d, err := db.Connect()
	if err != nil {
		return
	}

	var (
		r clients.RestClient
		w clients.WsClient
	)
	switch config.DataProvider {
	case "huobi":
		if r, err = huobi.Init(); err != nil {
			return
		}
		if w, err = huobi.InitWs(d.DataPipe()); err != nil {
			return
		}
	default:
		if r, err = cryptocompare.Init(); err != nil {
			return
		}
		if w, err = cryptocompare.InitWs(d.DataPipe()); err != nil {
			return
		}
	}
	p := clients.NewPuller(r, s, d.DataPipe())

	if err = p.RestoreLastSession(); err != nil {
		log.Println("error restoring last session")
	}

	// health checks
	e.GET("/healthz", SendOK)

	// serve web page
	e.LoadHTMLFiles("site/index.tmpl")
	e.Static("/css", "site/css")
	e.Static("/js", "site/js")
	e.GET("/", SendHTML)
	e.HEAD("/", SendOK)

	// serve api
	apiV1 := e.Group("/v1")
	{
		apiV1.POST("/collect/add", handlers.GinHandler(v1.AddWorker(p)))
		apiV1.GET("/collect/add", handlers.GinHandler(v1.AddWorker(p)))
		apiV1.POST("/collect/remove", handlers.GinHandler(v1.RemoveWorker(p)))
		apiV1.GET("/collect/remove", handlers.GinHandler(v1.RemoveWorker(p)))
		apiV1.GET("/collect/status", handlers.GinHandler(v1.PullingStatus(p, w)))
		apiV1.POST("/collect/update", handlers.GinHandler(v1.UpdateWorker(p)))
		apiV1.GET("/collect/update", handlers.GinHandler(v1.UpdateWorker(p)))
		apiV1.POST("/price", handlers.GinHandler(v1.Price(r, d)))
		apiV1.GET("/price", handlers.GinHandler(v1.Price(r, d)))
		apiV1.GET("/ws", ws.HandleWs(r, d))
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
