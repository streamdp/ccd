package validators

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/streamdp/ccd/config"
)

// Symbols - validate the field so that the value is from the list of currencies
func Symbols(fl validator.FieldLevel) bool {
	match := strings.Split(config.Symbols, ",")
	value := fl.Field().String()
	for i := range match {
		if match[i] == value {
			return true
		}
	}
	return false
}
