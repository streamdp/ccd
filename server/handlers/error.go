package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/pkg/wsclient"
)

// HandlerFuncResError to make router handler what return Result and error
type HandlerFuncResError func(*gin.Context) (*domain.Result, error)

var ErrBindQuery = errors.New("failed to bind query")

// GinHandler wrap HandlerFuncResError to easily handle and display errors nicely
func GinHandler(h HandlerFuncResError) gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h(c)
		if res == nil {
			res = &domain.Result{}
		}
		if res.Code == 0 {
			res.Code = getHttpStatus(err)
		}
		if err != nil {
			res.Message = err.Error()
			c.AbortWithStatusJSON(res.Code, res)

			return
		}
		c.JSON(res.Code, res)
	}
}

func getHttpStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, ErrBindQuery) ||
		errors.Is(err, wsclient.ErrNotSubscribed) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
