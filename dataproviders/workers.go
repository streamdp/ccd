package dataproviders

import (
	"github.com/streamdp/ccd/handlers"
	"time"
)

type Worker struct {
	Pipe chan *DataPipe  `json:"-"`
	done chan interface{}
	From string          `json:"from"`
	To       string          `json:"to"`
	Interval int   `json:"interval"`
}

type Workers struct {
	workers    map[*Worker]bool
	pipe       chan *DataPipe
	register   chan *Worker
	unregister chan *Worker
	dp         DataProvider
}

func NewWorkersControl(dp DataProvider) *Workers {
	return &Workers{
		workers:    make(map[*Worker]bool),
		pipe:       make(chan *DataPipe, 20),
		register:   make(chan *Worker),
		unregister: make(chan *Worker),
		dp:         dp,
	}
}

func (wc *Workers) NewWorker(from string, to string) *Worker {
	return &Worker{
		Pipe:     wc.pipe,
		done:     make(chan interface{}),
		From:     from,
		To:       to,
		Interval: 60,
	}
}

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

func (wc *Workers) GetPipe() chan *DataPipe {
	return wc.pipe
}

func (wc *Workers) GetWorkers() *map[*Worker]bool {
	return &wc.workers
}

func (wc *Workers) GetDataProvider() *DataProvider {
	return &wc.dp
}

func (wc *Workers) GetWorker(from string, to string) *Worker {
	for worker := range wc.workers {
		if worker.From == from && worker.To == to {
			return worker
		}
	}
	return nil
}

func (wc *Workers) Add(worker *Worker) *Worker {
	wc.register <- worker
	return worker
}

func (wc *Workers) AddWorker(from string, to string) *Worker {
	worker := wc.NewWorker(from, to)
	wc.register <- worker
	return worker
}

func (wc *Workers) RemoveWorker(from string, to string) {
	worker := wc.GetWorker(from, to)
	wc.unregister <- worker
}

func (w *Worker) Shutdown() {
	w.done <- 0
}

func (w *Worker) GetDone() chan interface{} {
	return w.done
}

func (w *Worker) Work(dp *DataProvider) {
	defer close(w.done)
	for {
		select {
		case <-w.done:
			return
		case <-time.After(time.Duration(w.Interval) * time.Second):
			data, err := (*dp).GetData(w.From, w.To)
			if err != nil {
				handlers.SystemHandler(err)
			}
			if data == GetEmptyData(w.From, w.To) {
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
