package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/domain"
)

// HandlerFuncResError to make router handler what return Result and error
type HandlerFuncResError func(*gin.Context) (domain.Result, error)

var ErrBindQuery = errors.New("failed to bind query")

// GinHandler wrap HandlerFuncResError to easily handle and display errors nicely
func GinHandler(h HandlerFuncResError) gin.HandlerFunc {
	return func(c *gin.Context) {
		if res, err := h(c); err != nil {
			res.Message = err.Error()
			if errors.Is(err, ErrBindQuery) {
				res.Code = http.StatusBadRequest
				c.AbortWithStatusJSON(res.Code, res)

				return
			}
			res.Code = http.StatusInternalServerError
			c.AbortWithStatusJSON(http.StatusInternalServerError, res)
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}
