package clients

import (
	"time"

	"github.com/streamdp/ccd/handlers"
)

// Task does all the data mining run
type Task struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int    `json:"interval"`
	done     chan struct{}
}
type Tasks map[string]*Task

func (t *Task) run(r RestClient, dataPipe chan *Data) {
	go func() {
		defer close(t.done)
		for {
			select {
			case <-t.done:
				return
			case <-time.After(time.Duration(t.Interval) * time.Second):
				data, err := r.Get(t.From, t.To)
				if err != nil {
					handlers.SystemHandler(err)
					continue
				}
				dataPipe <- data
			}
		}
	}()
}

func (t *Task) close() {
	t.done <- struct{}{}
}
