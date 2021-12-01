package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/streamdp/ccd/config"
	"strings"
)

func Common(fl validator.FieldLevel) bool {
	match := strings.Split(config.Common, ",")
	value := fl.Field().String()
	for _, s := range match {
		if s == value {
			return true
		}
	}
	return false
}
