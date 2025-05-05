package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SendHTML show a beautiful page with small intro and instruction
func SendHTML(appVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"year":    time.Now().Year(),
			"version": appVersion,
		})
	}
}

// SendOK using for HEAD request and send 200 and nil body
func SendOK(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
