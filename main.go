package main

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
	"github.com/streamdp/ccd/router"
)

func main() {
	config.ParseFlags()
	gin.SetMode(config.RunMode)
	engine := gin.Default()
	if err := engine.SetTrustedProxies(nil); err != nil {
		handlers.SystemHandler(err)
	}
	if err := router.InitRouter(engine); err != nil {
		handlers.SystemHandler(err)
		return
	}
	if err := engine.Run(config.Port); err != nil {
		handlers.SystemHandler(err)
		return
	}
}
