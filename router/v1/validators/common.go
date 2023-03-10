package validators

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/streamdp/ccd/config"
)

// Common - validate the field so that the value is from the list of common currencies
func Common(fl validator.FieldLevel) bool {
	match := strings.Split(config.Common, ",")
	value := fl.Field().String()
	for i := range match {
		if match[i] == value {
			return true
		}
	}
	return false
}
