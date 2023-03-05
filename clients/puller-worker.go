package clients

import (
	"time"

	"github.com/streamdp/ccd/handlers"
)

// Worker does all the data mining run
type Worker struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int    `json:"interval"`
	done     chan struct{}
}
type Workers map[*Worker]struct{}

func (w *Worker) run(r RestClient, dataPipe chan *Data) {
	go func() {
		defer close(w.done)
		for {
			select {
			case <-w.done:
				return
			case <-time.After(time.Duration(w.Interval) * time.Second):
				data, err := r.Get(w.From, w.To)
				if err != nil {
					handlers.SystemHandler(err)
					continue
				}
				dataPipe <- data
			}
		}
	}()
}

func (w *Worker) stop() {
	w.done <- struct{}{}
}
