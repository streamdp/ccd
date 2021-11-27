package dataproviders

import (
	"sync"
)

type Worker struct {
	pipe    chan *DataPipe
	done    chan interface{}
	isAlive bool
}

type Workers struct {
	list  map[string]map[string]*Worker
	mutex *sync.Mutex
}

func (w *Workers) GetWorker(from, to string) *Worker {
	if w.list[from] == nil || w.list[from][to] == nil {
		return nil
	} else {
		return w.list[from][to]
	}
}

func (w *Workers) AddWorker(from string, to string, pipe chan *DataPipe) *Worker {
	w.mutex.Lock()
	if w.list[from] == nil {
		w.list[from] = make(map[string]*Worker)
		w.list[from][to] = CreateWorker(pipe)
	} else if w.list[from][to] == nil {
		w.list[from][to] = CreateWorker(pipe)
	}
	w.mutex.Unlock()
	return w.list[from][to]
}

func CreateWorkersControl() *Workers {
	return &Workers{
		list:  make(map[string]map[string]*Worker),
		mutex: new(sync.Mutex),
	}
}

func CreateWorker(pipe chan *DataPipe) *Worker {
	return &Worker{
		pipe:    pipe,
		done:    make(chan interface{}),
		isAlive: false,
	}
}

func (w *Worker) Shutdown() {
	w.done <- 0
}

func (w *Worker) IsAlive() (result bool) {
	if w != nil {
		result = w.isAlive
	}
	return
}

func (w *Worker) SetAlive(alive bool) {
	w.isAlive = alive
}

func (w *Worker) GetDone() chan interface{} {
	return w.done
}

func (w *Worker) GetPipe() chan *DataPipe {
	return w.pipe
}
