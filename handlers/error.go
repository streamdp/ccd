package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// HandlerFuncResError to make router handler what return Result and error
type HandlerFuncResError func(*gin.Context) (Result, error)

// GinHandler wrap HandlerFuncResError to easily handle and display errors nicely
func GinHandler(myHandler HandlerFuncResError) gin.HandlerFunc {
	return func(c *gin.Context) {
		if res, err := myHandler(c); err != nil {
			res.UpdateAllFields(http.StatusInternalServerError, err.Error(), nil)
			c.AbortWithStatusJSON(http.StatusInternalServerError, res)
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}

// SystemHandler handling and logging error
func SystemHandler(err error) {
	log.Println(err)
}
