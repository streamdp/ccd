package router

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func ServeHTML(c *gin.Context)  {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"time": time.Now().UTC().String(),
	})
}