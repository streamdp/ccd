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
			res.SetCode(http.StatusInternalServerError)
			res.SetMessage(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, res)
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}

func SystemHandler(err error) {
	log.Println(err)
}
