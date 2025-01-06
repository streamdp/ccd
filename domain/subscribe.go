package domain

import (
	"strings"
)

type Subscription struct {
	From string `json:"from"`
	To   string `json:"to"`
	id   int64
}
type Subscriptions map[string]*Subscription

func NewSubscription(from, to string, id int64) *Subscription {
	return &Subscription{
		From: strings.ToUpper(from),
		To:   strings.ToUpper(to),
		id:   id,
	}
}

func (s *Subscription) Id() int64 {
	return s.id
}
