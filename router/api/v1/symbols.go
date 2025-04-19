package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccd/router/handlers"
)

type SymbolsRepo interface {
	Update(symbol, unicode string) error
	Load() error
	GetAll() []string
	Add(symbol, unicode string) error
	Remove(symbol string) error
	IsPresent(symbol string) bool
}

// SymbolQuery structure for easily json serialization/validation/binding GET and POST query data
type SymbolQuery struct {
	Symbol  string `binding:"required" form:"symbol"  json:"symbol"`
	Unicode string `form:"unicode"     json:"unicode"`
}

// AllSymbols return all symbols
func AllSymbols(sr SymbolsRepo) handlers.HandlerFuncResError {
	return func(c *gin.Context) (handlers.Result, error) {
		return handlers.Result{}.UpdateAllFields(
			http.StatusOK, "list of all symbols presented", sr.GetAll(),
		), nil
	}
}

// AddSymbol to the symbols table
func AddSymbol(sr SymbolsRepo) handlers.HandlerFuncResError {
	return func(c *gin.Context) (handlers.Result, error) {
		q := SymbolQuery{}
		if err := c.Bind(&q); err != nil {
			return handlers.Result{}, err
		}
		if err := sr.Add(q.Symbol, q.Unicode); err != nil {
			return handlers.Result{}, err
		}

		return handlers.Result{}.UpdateAllFields(
			http.StatusOK, fmt.Sprintf("symbol %s successfully added to the db", q.Symbol), nil,
		), nil
	}
}

// UpdateSymbol in the symbols table
func UpdateSymbol(sr SymbolsRepo) handlers.HandlerFuncResError {
	return func(c *gin.Context) (handlers.Result, error) {
		q := SymbolQuery{}
		if err := c.Bind(&q); err != nil {
			return handlers.Result{}, err
		}
		if err := sr.Update(q.Symbol, q.Unicode); err != nil {
			return handlers.Result{}, err
		}

		return handlers.Result{}.UpdateAllFields(
			http.StatusOK, fmt.Sprintf("symbol %s successfully updated", q.Symbol), nil,
		), nil
	}
}

// RemoveSymbol from the symbols table
func RemoveSymbol(sr SymbolsRepo) handlers.HandlerFuncResError {
	return func(c *gin.Context) (handlers.Result, error) {
		q := SymbolQuery{}
		if err := c.Bind(&q); err != nil {
			return handlers.Result{}, err
		}
		if err := sr.Remove(q.Symbol); err != nil {
			return handlers.Result{}, err
		}

		return handlers.Result{}.UpdateAllFields(
			http.StatusOK, fmt.Sprintf("symbol %s successfully removed", q.Symbol), nil,
		), nil
	}
}

// ValidateSymbols Symbols - validate the field so that the value is from the list of currencies
func ValidateSymbols(sr SymbolsRepo) func(fl validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		return sr.IsPresent(fl.Field().String())
	}
}
