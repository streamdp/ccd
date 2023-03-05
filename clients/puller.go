package clients

import (
	"sync"

	"github.com/streamdp/ccd/config"
)

// RestPuller puller base struct
type RestPuller struct {
	w        Workers
	dataPipe chan *Data
	client   RestClient
	pullerMu sync.RWMutex
}

// NewPuller init rest puller
func NewPuller(r RestClient, dataPipe chan *Data) *RestPuller {
	return &RestPuller{
		w:        Workers{},
		dataPipe: dataPipe,
		client:   r,
	}
}

func (p *RestPuller) createWorker(from string, to string, interval int) *Worker {
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

// ListWorkers return all puller w
func (p *RestPuller) ListWorkers() Workers {
	var w = Workers{}
	p.pullerMu.RLock()
	for k := range p.w {
		w[k] = struct{}{}
	}
	p.pullerMu.RUnlock()
	return w
}

// Client return rest client
func (p *RestPuller) Client() RestClient {
	return p.client
}

// Worker return worker for the selected currencies pair, if possible
func (p *RestPuller) Worker(from string, to string) *Worker {
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()
	for w := range p.w {
		if w.From == from && w.To == to {
			return w
		}
	}
	return nil
}

// AddWorker to collect data for the selected currency pair to the puller
func (p *RestPuller) AddWorker(from string, to string, interval int) *Worker {
	w := p.createWorker(from, to, interval)
	w.run(p.Client(), p.DataPipe())
	p.pullerMu.Lock()
	p.w[w] = struct{}{}
	p.pullerMu.Unlock()
	return w
}

// RemoveWorker from the puller by the selected currency pair
func (p *RestPuller) RemoveWorker(from string, to string) {
	w := p.Worker(from, to)
	w.stop()
	p.pullerMu.Lock()
	defer p.pullerMu.Unlock()
	delete(p.w, w)
}
