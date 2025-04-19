package config

import (
	"errors"
	"fmt"
	"time"
)

const (
	httpServerDefaultPort = 8080
	httpDefaultTimeout    = 5000
)

var errWrongNetworkPort = errors.New("port must be between 0 and 65535")

type Http struct {
	port          int
	clientTimeout int
}

func (h *Http) ClientTimeout() time.Duration {
	return time.Duration(h.clientTimeout) * time.Millisecond
}

func (h *Http) Port() int {
	return h.port
}

func (h *Http) Validate() error {
	if h.port < 0 || h.port > 65535 {
		return fmt.Errorf("http: %w", errWrongNetworkPort)
	}

	return nil
}
