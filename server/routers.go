package server

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	v1 "github.com/streamdp/ccd/server/api/v1"
	"github.com/streamdp/ccd/server/handlers"
)

// InitRouter basic work on setting up the application, declare endpoints, register our custom validation functions
func (s *Server) InitRouter(ctx context.Context) error {
	// health checks
	s.GET("/healthz", SendOK)
	s.HEAD("/healthz", SendOK)

	// serve web page
	s.LoadHTMLFiles("site/index.tmpl")
	s.Static("/css", "site/css")
	s.Static("/js", "site/js")
	s.GET("/", SendHTML(s.cfg.Version()))
	s.HEAD("/", SendOK)

	// DEPRECATED: use v2 api instead
	apiV1 := s.Group("/v1")
	{
		apiV1.GET("/collect/status", handlers.GinHandler(v1.PullingStatus(s.p, s.wc)))
		apiV1.GET("/collect/add", handlers.GinHandler(v1.AddWorker(ctx, s.p)))
		apiV1.POST("/collect", handlers.GinHandler(v1.AddWorker(ctx, s.p)))
		apiV1.GET("/collect/remove", handlers.GinHandler(v1.RemoveWorker(ctx, s.p)))
		apiV1.DELETE("/collect", handlers.GinHandler(v1.RemoveWorker(ctx, s.p)))
		apiV1.GET("/collect/update", handlers.GinHandler(v1.UpdateWorker(ctx, s.p)))
		apiV1.PUT("/collect", handlers.GinHandler(v1.UpdateWorker(ctx, s.p)))

		apiV1.GET("/symbols", handlers.GinHandler(v1.AllSymbols(s.sr)))
		apiV1.GET("/symbols/add", handlers.GinHandler(v1.AddSymbol(s.sr)))
		apiV1.POST("/symbols", handlers.GinHandler(v1.AddSymbol(s.sr)))
		apiV1.GET("/symbols/update", handlers.GinHandler(v1.UpdateSymbol(s.sr)))
		apiV1.PUT("/symbols", handlers.GinHandler(v1.UpdateSymbol(s.sr)))
		apiV1.GET("/symbols/remove", handlers.GinHandler(v1.RemoveSymbol(s.sr)))
		apiV1.DELETE("/symbols", handlers.GinHandler(v1.RemoveSymbol(s.sr)))

		apiV1.GET("/price", handlers.GinHandler(v1.Price(s.rc, s.d)))
		apiV1.POST("/price", handlers.GinHandler(v1.Price(s.rc, s.d)))

		apiV1.GET("/ws", v1.HandleWs(ctx, s.ws))
		if s.wc != nil {
			apiV1.GET("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, s.wc)))
			apiV1.POST("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, s.wc)))
			apiV1.GET("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, s.wc)))
			apiV1.POST("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, s.wc)))
		}
	}

	// actual version of API
	apiV2 := s.Group("/v2")
	{
		// collect
		apiV2.GET("/collect", handlers.GinHandler(v1.PullingStatus(s.p, s.wc)))
		apiV2.POST("/collect", handlers.GinHandler(v1.AddWorker(ctx, s.p)))
		apiV2.PUT("/collect", handlers.GinHandler(v1.UpdateWorker(ctx, s.p)))
		apiV2.DELETE("/collect", handlers.GinHandler(v1.RemoveWorker(ctx, s.p)))
		// symbols
		apiV2.GET("/symbols", handlers.GinHandler(v1.AllSymbols(s.sr)))
		apiV2.POST("/symbols", handlers.GinHandler(v1.AddSymbol(s.sr)))
		apiV2.PUT("/symbols", handlers.GinHandler(v1.UpdateSymbol(s.sr)))
		apiV2.DELETE("/symbols", handlers.GinHandler(v1.RemoveSymbol(s.sr)))
		// price
		apiV2.GET("/price", handlers.GinHandler(v1.Price(s.rc, s.d)))
		// websockets
		apiV2.GET("/ws", v1.HandleWs(ctx, s.ws))
		if s.wc != nil {
			apiV2.GET("/ws/subscribe", handlers.GinHandler(v1.Subscribe(ctx, s.wc)))
			apiV2.GET("/ws/unsubscribe", handlers.GinHandler(v1.Unsubscribe(ctx, s.wc)))
		}
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("symbols", v1.ValidateSymbols(s.sr)); err != nil {
			return fmt.Errorf("failed to register validator: %w", err)
		}
	}

	return nil
}
