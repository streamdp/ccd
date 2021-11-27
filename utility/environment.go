package utility

import (
	"os"
	"strings"
)

func GetEnv(name string) (result string)  {
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if pair[0] == name {
			return pair[1]
		}
	}
	return
}
