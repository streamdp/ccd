package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/handlers"
	"net/http"
)

func Ping(_ *gin.Context) (res handlers.Result, err error) {
	res.UpdateAllFields(http.StatusOK, "pong", nil)
	return
}
