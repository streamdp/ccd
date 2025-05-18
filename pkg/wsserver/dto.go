package ws

import (
	"fmt"
	"strings"
)

type wsMessage struct {
	T      string `json:"type"`
	Pair   pair   `json:"pair"`
	Reason string `json:"reason,omitempty"`
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
