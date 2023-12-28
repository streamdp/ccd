package main

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
	"github.com/streamdp/ccd/router"
	"github.com/streamdp/ccd/session"
)

func main() {
	config.ParseFlags()
	gin.SetMode(config.RunMode)
	e := gin.Default()
	if err := router.InitRouter(e, session.NewKeysStore()); err != nil {
		handlers.SystemHandler(err)
		return
	}
	if err := e.Run(config.Port); err != nil {
		handlers.SystemHandler(err)
		return
	}
}
