package v1

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/handlers"
)

// SymbolQuery structure for easily json serialization/validation/binding GET and POST query data
type SymbolQuery struct {
	Symbol  string `json:"symbol" form:"symbol" binding:"required"`
	Unicode string `json:"unicode" form:"unicode" binding:"required"`
}

// AddSymbol to the symbols table
func AddSymbol(db db.Database) handlers.HandlerFuncResError {
	return func(c *gin.Context) (r handlers.Result, err error) {
		q := SymbolQuery{}
		if err = c.Bind(&q); err != nil {
			return
		}
		if _, err = db.AddSymbol(strings.ToUpper(q.Symbol), strings.ToUpper(q.Unicode)); err != nil {
			return
		}
		r.UpdateAllFields(http.StatusOK, fmt.Sprintf("symbol %s successfully added to the db", q.Symbol), nil)
		return
	}
}

// UpdateSymbol in the symbols table
func UpdateSymbol(db db.Database) handlers.HandlerFuncResError {
	return func(c *gin.Context) (r handlers.Result, err error) {
		q := SymbolQuery{}
		if err = c.Bind(&q); err != nil {
			return
		}
		if _, err = db.UpdateSymbol(strings.ToUpper(q.Symbol), strings.ToUpper(q.Unicode)); err != nil {
			return
		}
		r.UpdateAllFields(http.StatusOK, fmt.Sprintf("symbol %s successfully updated", q.Symbol), nil)
		return
	}
}

// RemoveSymbol from the symbols table
func RemoveSymbol(db db.Database) handlers.HandlerFuncResError {
	return func(c *gin.Context) (r handlers.Result, err error) {
		q := SymbolQuery{}
		if err = c.Bind(&q); err != nil {
			return
		}
		if _, err = db.RemoveSymbol(strings.ToUpper(q.Symbol)); err != nil {
			return
		}
		r.UpdateAllFields(http.StatusOK, fmt.Sprintf("symbol %s successfully removed", q.Symbol), nil)
		return
	}
}
