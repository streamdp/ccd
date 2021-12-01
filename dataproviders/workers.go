package dataproviders

import (
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
	workers    map[*Worker]bool
	pipe       chan *DataPipe
	register   chan *Worker
	unregister chan *Worker
	dp         DataProvider
}

// NewWorkersControl init Workers structure
func NewWorkersControl(dp DataProvider) *Workers {
	return &Workers{
		workers:    make(map[*Worker]bool),
		pipe:       make(chan *DataPipe, 20),
		register:   make(chan *Worker),
		unregister: make(chan *Worker),
		dp:         dp,
	}
}

// NewWorker init Worker structure with selected currencies pair
func (wc *Workers) NewWorker(from string, to string) *Worker {
	return &Worker{
		Pipe:     wc.pipe,
		done:     make(chan interface{}),
		From:     from,
		To:       to,
		Interval: 60,
	}
}

// Run managing Worker
func (wc *Workers) Run() {
	for {
		select {
		case worker := <-wc.register:
			wc.workers[worker] = true
		case worker := <-wc.unregister:
			wc.workers[worker] = false
			worker.done <- 0
			delete(wc.workers, worker)
		}
	}
}

// GetPipe return common DataPipe
func (wc *Workers) GetPipe() chan *DataPipe {
	return wc.pipe
}

// GetWorkers return state of the all *Workers
func (wc *Workers) GetWorkers() *map[*Worker]bool {
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

// Add Worker to the managing service by pointer
func (wc *Workers) Add(worker *Worker) *Worker {
	wc.register <- worker
	return worker
}

// AddWorker a new worker that will collect data for the selected currency pair to the management service
func (wc *Workers) AddWorker(from string, to string) *Worker {
	worker := wc.NewWorker(from, to)
	wc.Add(worker)
	return worker
}

//RemoveWorker from the managing service by the selected currency pair
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
}
