package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/handlers"
	"net/http"
)

// Ping for check if service is alive
func Ping(_ *gin.Context) (res handlers.Result, err error) {
	res.UpdateAllFields(http.StatusOK, "pong", nil)
	return
}
