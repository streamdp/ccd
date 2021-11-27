package dataproviders

import (
	"sync"
)

type Worker struct {
	pipe    chan *DataPipe
	done    chan interface{}
	IsAlive bool `json:"is_alive"`
}

type Workers struct {
	List  map[string]map[string]*Worker `json:"workers"`
	mutex *sync.Mutex
}

func (w *Workers) GetWorker(from, to string) *Worker {
	if w.List[from] == nil || w.List[from][to] == nil {
		return nil
	} else {
		return w.List[from][to]
	}
}

func (w *Workers) AddWorker(from string, to string, pipe chan *DataPipe) *Worker {
	w.mutex.Lock()
	if w.List[from] == nil {
		w.List[from] = make(map[string]*Worker)
		w.List[from][to] = CreateWorker(pipe)
	} else if w.List[from][to] == nil {
		w.List[from][to] = CreateWorker(pipe)
	} else if !w.List[from][to].IsAlive {
		w.List[from][to] = CreateWorker(pipe)
	}
	w.mutex.Unlock()
	return w.List[from][to]
}

func CreateWorkersControl() *Workers {
	return &Workers{
		List:  make(map[string]map[string]*Worker),
		mutex: new(sync.Mutex),
	}
}

func CreateWorker(pipe chan *DataPipe) *Worker {
	return &Worker{
		pipe:    pipe,
		done:    make(chan interface{}),
		IsAlive: false,
	}
}

func (w *Worker) Shutdown() {
	w.done <- 0
}

func (w *Worker) SetAlive(alive bool) {
	w.IsAlive = alive
}

func (w *Worker) GetDone() chan interface{} {
	return w.done
}

func (w *Worker) GetPipe() chan *DataPipe {
	return w.pipe
}
