package router

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/config"
	"net/http"
	"time"
)

func IndexHTML(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"year":    time.Now().Year(),
		"version": config.Version,
	})
}
