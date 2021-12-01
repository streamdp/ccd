package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccd/config"
	"strings"
)

// Crypto - validate the field so that the value is from the list of сrypto currencies
func Crypto(fl validator.FieldLevel) bool {
	match := strings.Split(config.Crypto, ",")
	value := fl.Field().String()
	for _, s := range match {
		if s == value {
			return true
		}
	}
	return false
}
