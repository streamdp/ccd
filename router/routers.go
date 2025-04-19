package router

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/db"
	v1 "github.com/streamdp/ccd/router/api/v1"
	"github.com/streamdp/ccd/router/api/v1/ws"
	"github.com/streamdp/ccd/router/handlers"
)

// InitRouter basic work on setting up the application, declare endpoints, register our custom validation functions
func InitRouter(
	ctx context.Context,
	e *gin.Engine,
	d db.Database,
	l *log.Logger,
	sr v1.SymbolsRepo,
	r clients.RestClient,
	w clients.WsClient,
	p clients.RestApiPuller,
) error {
	// health checks
	e.GET("/healthz", SendOK)
	e.HEAD("/healthz", SendOK)

	// serve web page
	e.LoadHTMLFiles("site/index.tmpl")
	e.Static("/css", "site/css")
	e.Static("/js", "site/js")
	e.GET("/", SendHTML)
	e.HEAD("/", SendOK)

	// DEPRECATED: use v2 api instead
	apiV1 := e.Group("/v1")
	{
		apiV1.GET("/collect/status", handlers.GinHandler(v1.PullingStatus(p, w)))
		apiV1.GET("/collect/add", handlers.GinHandler(v1.AddWorker(p)))
		apiV1.POST("/collect", handlers.GinHandler(v1.AddWorker(p)))
		apiV1.GET("/collect/remove", handlers.GinHandler(v1.RemoveWorker(p)))
		apiV1.DELETE("/collect", handlers.GinHandler(v1.RemoveWorker(p)))
		apiV1.GET("/collect/update", handlers.GinHandler(v1.UpdateWorker(p)))
		apiV1.PUT("/collect", handlers.GinHandler(v1.UpdateWorker(p)))

		apiV1.GET("/symbols", handlers.GinHandler(v1.AllSymbols(sr)))
		apiV1.GET("/symbols/add", handlers.GinHandler(v1.AddSymbol(sr)))
		apiV1.POST("/symbols", handlers.GinHandler(v1.AddSymbol(sr)))
		apiV1.GET("/symbols/update", handlers.GinHandler(v1.UpdateSymbol(sr)))
		apiV1.PUT("/symbols", handlers.GinHandler(v1.UpdateSymbol(sr)))
		apiV1.GET("/symbols/remove", handlers.GinHandler(v1.RemoveSymbol(sr)))
		apiV1.DELETE("/symbols", handlers.GinHandler(v1.RemoveSymbol(sr)))

		apiV1.GET("/price", handlers.GinHandler(v1.Price(r, d)))
		apiV1.POST("/price", handlers.GinHandler(v1.Price(r, d)))

		apiV1.GET("/ws", ws.HandleWs(ctx, r, l, d))
		if w != nil {
			apiV1.GET("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, w)))
			apiV1.POST("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, w)))
			apiV1.GET("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, w)))
			apiV1.POST("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, w)))
		}
	}

	// actual version of API
	apiV2 := e.Group("/v2")
	{
		// collect
		apiV2.GET("/collect", handlers.GinHandler(v1.PullingStatus(p, w)))
		apiV2.POST("/collect", handlers.GinHandler(v1.AddWorker(p)))
		apiV2.PUT("/collect", handlers.GinHandler(v1.UpdateWorker(p)))
		apiV2.DELETE("/collect", handlers.GinHandler(v1.RemoveWorker(p)))
		// symbols
		apiV2.GET("/symbols", handlers.GinHandler(v1.AllSymbols(sr)))
		apiV2.POST("/symbols", handlers.GinHandler(v1.AddSymbol(sr)))
		apiV2.PUT("/symbols", handlers.GinHandler(v1.UpdateSymbol(sr)))
		apiV2.DELETE("/symbols", handlers.GinHandler(v1.RemoveSymbol(sr)))
		// price
		apiV2.GET("/price", handlers.GinHandler(v1.Price(r, d)))
		// websockets
		apiV2.GET("/ws", ws.HandleWs(ctx, r, l, d))
		if w != nil {
			apiV2.GET("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, w)))
			apiV2.GET("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, w)))
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("symbols", v1.ValidateSymbols(sr)); err != nil {
			return fmt.Errorf("failed to register validator: %w", err)
		}
	}

	return nil
}
