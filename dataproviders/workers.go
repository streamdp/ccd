package dataproviders

import (
	"sync"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
)

// Worker does all the data mining work
type Worker struct {
	Pipe     chan *DataPipe `json:"-"`
	done     chan interface{}
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int    `json:"interval"`
}

// Workers this is a manager who manages a group of Worker
type Workers struct {
	workers map[*Worker]struct{}
	pipe    chan *DataPipe
	dp      DataProvider
	mu      sync.RWMutex
}

// NewWorkersControl init Workers structure
func NewWorkersControl(dp DataProvider) *Workers {
	return &Workers{
		workers: make(map[*Worker]struct{}),
		pipe:    make(chan *DataPipe, 20),
		dp:      dp,
	}
}

// NewWorker init Worker structure with selected currencies pair
func (wc *Workers) NewWorker(from string, to string, interval int) *Worker {
	if interval <= 0 {
		interval = config.DefaultPullingInterval
	}
	return &Worker{
		Pipe:     wc.pipe,
		done:     make(chan interface{}),
		From:     from,
		To:       to,
		Interval: interval,
	}
}

// GetPipe return common DataPipe
func (wc *Workers) GetPipe() chan *DataPipe {
	return wc.pipe
}

// GetWorkers return state of the all *Workers
func (wc *Workers) GetWorkers() map[*Worker]struct{} {
	var w = map[*Worker]struct{}{}
	wc.mu.RLock()
	for k := range wc.workers {
		w[k] = struct{}{}
	}
	wc.mu.RUnlock()
	return w
}

// GetDataProvider return *DataProvider
func (wc *Workers) GetDataProvider() *DataProvider {
	return &wc.dp
}

// GetWorker for the selected currencies pair, if possible
func (wc *Workers) GetWorker(from string, to string) *Worker {
	wc.mu.RLock()
	defer wc.mu.RUnlock()
	for worker := range wc.workers {
		if worker.From == from && worker.To == to {
			return worker
		}
	}
	return nil
}

// AddWorker a new worker that will collect data for the selected currency pair to the management service
func (wc *Workers) AddWorker(from string, to string, interval int) *Worker {
	worker := wc.NewWorker(from, to, interval)
	wc.mu.Lock()
	wc.workers[worker] = struct{}{}
	wc.mu.Unlock()
	return worker
}

// RemoveWorker from the managing service by the selected currency pair
func (wc *Workers) RemoveWorker(from string, to string) {
	worker := wc.GetWorker(from, to)
	worker.Close()
	wc.mu.Lock()
	defer wc.mu.Unlock()
	delete(wc.workers, worker)
}

// Close Worker
func (w *Worker) Close() {
	w.done <- 0
}

// Work of the Worker is collect Data and send it throughout the Pipe
func (w *Worker) Work(dp *DataProvider) {
	go func() {
		defer close(w.done)
		for {
			select {
			case <-w.done:
				return
			case <-time.After(time.Duration(w.Interval) * time.Second):
				data, err := (*dp).Get(w.From, w.To)
				if err != nil {
					handlers.SystemHandler(err)
					continue
				}
				w.Pipe <- &DataPipe{
					From: w.From,
					To:   w.To,
					Data: data,
				}
			}
		}
	}()
}
