package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

func Ping(_ *gin.Context) (res handlers.Result, err error) {
	res.UpdateAllFields(http.StatusOK, "pong", nil)
	return
}
