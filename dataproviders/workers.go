package dataproviders

import (
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
	"time"
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
	workers    map[*Worker]struct{}
	pipe       chan *DataPipe
	unregister chan *Worker
	dp         DataProvider
}

// NewWorkersControl init Workers structure
func NewWorkersControl(dp DataProvider) *Workers {
	return &Workers{
		workers:    make(map[*Worker]struct{}),
		pipe:       make(chan *DataPipe, 20),
		unregister: make(chan *Worker),
		dp:         dp,
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

// Run managing Workers
func (wc *Workers) Run() {
	go func() {
		for worker := range wc.unregister {
			worker.done <- 0
			delete(wc.workers, worker)
		}
	}()
}

// GetPipe return common DataPipe
func (wc *Workers) GetPipe() chan *DataPipe {
	return wc.pipe
}

// GetWorkers return state of the all *Workers
func (wc *Workers) GetWorkers() *map[*Worker]struct{} {
	return &wc.workers
}

// GetDataProvider return *DataProvider
func (wc *Workers) GetDataProvider() *DataProvider {
	return &wc.dp
}

// GetWorker for the selected currencies pair, if possible
func (wc *Workers) GetWorker(from string, to string) *Worker {
	for worker := range wc.workers {
		if worker.From == from && worker.To == to {
			return worker
		}
	}
	return nil
}

// Register Worker to the managing service
func (wc *Workers) Register(worker *Worker) *Worker {
	wc.workers[worker] = struct{}{}
	return worker
}

// AddWorker a new worker that will collect data for the selected currency pair to the management service
func (wc *Workers) AddWorker(from string, to string, interval int) *Worker {
	worker := wc.NewWorker(from, to, interval)
	wc.Register(worker)
	return worker
}

// RemoveWorker from the managing service by the selected currency pair
func (wc *Workers) RemoveWorker(from string, to string) {
	worker := wc.GetWorker(from, to)
	wc.unregister <- worker
}

// Shutdown Worker
func (w *Worker) Shutdown() {
	w.done <- 0
}

// GetDone return chan interface{} for the selected Worker by pointer
func (w *Worker) GetDone() chan interface{} {
	return w.done
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
