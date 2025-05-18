package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/streamdp/ccd/domain"
)

type client struct {
	handler *handler
	cancel  context.CancelFunc
}

type wsMessage struct {
	T       string       `json:"type"`
	Pair    *pair        `json:"pair,omitempty"`
	Data    *domain.Data `json:"data,omitempty"`
	Message string       `json:"message,omitempty"`
}

func (w *wsMessage) Bytes() []byte {
	b, _ := json.Marshal(w)

	return b
}

type pair struct {
	From string `json:"fsym"`
	To   string `json:"tsym"`
}

func (p *pair) toUpper() {
	p.From = strings.ToUpper(p.From)
	p.To = strings.ToUpper(p.To)
}

func (p *pair) buildName() string {
	return fmt.Sprintf("%s:%s", p.From, p.To)
}
