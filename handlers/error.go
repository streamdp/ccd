package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type HandlerFuncResError func(*gin.Context) (Result, error)

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

func SystemHandler(err error) {
	log.Println(err)
}
