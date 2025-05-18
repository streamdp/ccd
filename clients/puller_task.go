package clients

import (
	"log"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/streamdp/ccd/domain"
)

const defaultRunTaskGap = 30

// Task does all the data mining run
type Task struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int64  `json:"interval"`
	done     chan struct{}
}
type Tasks map[string]*Task

func (t *Task) run(r RestClient, l *log.Logger, dataPipe []chan *domain.Data) {
	timer := time.NewTimer(time.Duration(rand.Intn(defaultRunTaskGap)) * time.Second)
	go func() {
		defer close(t.done)
		for {
			select {
			case <-t.done:
				timer.Stop()

				return
			case <-timer.C:
				timer.Reset(time.Duration(atomic.LoadInt64(&t.Interval)) * time.Second)
				data, err := r.Get(t.From, t.To)
				if err != nil {
					l.Println(err)

					continue
				}
				for i := range dataPipe {
					dataPipe[i] <- data
				}
			}
		}
	}()
}

func (t *Task) close() {
	t.done <- struct{}{}
}
