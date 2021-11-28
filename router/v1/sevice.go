package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/streamdp/ccdatacollector/entity"
	"github.com/streamdp/ccdatacollector/handlers"
	"net/http"
)

func Ping() handlers.HandlerFuncResError {
	return func(c *gin.Context) (res entity.Result, err error) {
		res.UpdateAllFields(http.StatusOK, "pong", nil)
		return
	}
}
