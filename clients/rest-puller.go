package clients

import (
	"sync"
	"time"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/handlers"
)

// Worker does all the data mining run
type Worker struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Interval int    `json:"interval"`
	done     chan struct{}
}

// RestPuller puller base struct
type RestPuller struct {
	workers  map[*Worker]struct{}
	dataPipe chan *Data
	client   RestClient
	pullerMu sync.RWMutex
}

// NewPuller init rest puller
func NewPuller(r RestClient, dataPipe chan *Data) *RestPuller {
	return &RestPuller{
		workers:  make(map[*Worker]struct{}),
		dataPipe: dataPipe,
		client:   r,
	}
}

func (p *RestPuller) newWorker(from string, to string, interval int) *Worker {
	if interval <= 0 {
		interval = config.DefaultPullingInterval
	}
	return &Worker{
		done:     make(chan struct{}),
		From:     from,
		To:       to,
		Interval: interval,
	}
}

// DataPipe return communication channel between pullers and database
func (p *RestPuller) DataPipe() chan *Data {
	return p.dataPipe
}

// Workers return all puller workers
func (p *RestPuller) Workers() map[*Worker]struct{} {
	var w = map[*Worker]struct{}{}
	p.pullerMu.RLock()
	for k := range p.workers {
		w[k] = struct{}{}
	}
	p.pullerMu.RUnlock()
	return w
}

// Client return rest client
func (p *RestPuller) Client() *RestClient {
	return &p.client
}

// Worker return worker for the selected currencies pair, if possible
func (p *RestPuller) Worker(from string, to string) *Worker {
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()
	for worker := range p.workers {
		if worker.From == from && worker.To == to {
			return worker
		}
	}
	return nil
}

// AddWorker to collect data for the selected currency pair to the puller
func (p *RestPuller) AddWorker(from string, to string, interval int) *Worker {
	worker := p.newWorker(from, to, interval)
	worker.run(p.Client(), p.DataPipe())
	p.pullerMu.Lock()
	p.workers[worker] = struct{}{}
	p.pullerMu.Unlock()
	return worker
}

// RemoveWorker from the puller by the selected currency pair
func (p *RestPuller) RemoveWorker(from string, to string) {
	worker := p.Worker(from, to)
	worker.close()
	p.pullerMu.Lock()
	defer p.pullerMu.Unlock()
	delete(p.workers, worker)
}

func (w *Worker) close() {
	w.done <- struct{}{}
}

func (w *Worker) run(dp *RestClient, dataPipe chan *Data) {
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
				dataPipe <- data
			}
		}
	}()
}
