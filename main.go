package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
	"github.com/streamdp/ccd/router"
	"github.com/streamdp/ccd/session"
)

func main() {
	config.ParseFlags()
	gin.SetMode(config.RunMode)
	s, err := session.NewKeysStore()
	if err != nil {
		log.Println(fmt.Errorf("failed to init session store: %w", err))
	}
	e := gin.Default()
	if err = router.InitRouter(e, s); err != nil {
		handlers.SystemHandler(err)
		return
	}
	if err = e.Run(config.Port); err != nil {
		handlers.SystemHandler(err)
		return
	}
}
