package clients

import (
	"strings"
	"sync"

	"github.com/streamdp/ccd/config"
)

// RestPuller puller base struct
type RestPuller struct {
	t        Tasks
	dataPipe chan *Data
	client   RestClient
	pullerMu sync.RWMutex
}

// NewPuller init rest puller
func NewPuller(r RestClient, dataPipe chan *Data) RestApiPuller {
	return &RestPuller{
		t:        Tasks{},
		dataPipe: dataPipe,
		client:   r,
	}
}

func (p *RestPuller) newTask(from string, to string, interval int) *Task {
	if interval <= 0 {
		interval = config.DefaultPullingInterval
	}
	return &Task{
		done:     make(chan struct{}),
		From:     from,
		To:       to,
		Interval: interval,
	}
}

// ListTasks return all tasks
func (p *RestPuller) ListTasks() Tasks {
	var t = make(Tasks, len(p.t))
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()
	for k, v := range p.t {
		t[k] = v
	}
	return t
}

// Task return task with selected currencies pair, if possible
func (p *RestPuller) Task(from string, to string) *Task {
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()
	return p.t[buildTaskName(from, to)]
}

func buildTaskName(from, to string) string {
	return strings.ToUpper(from + to)
}

// AddTask to collect data for the selected currency pair to the puller
func (p *RestPuller) AddTask(from string, to string, interval int) *Task {
	t := p.newTask(from, to, interval)
	t.run(p.client, p.dataPipe)
	p.pullerMu.Lock()
	p.t[buildTaskName(from, to)] = t
	p.pullerMu.Unlock()
	return t
}

// RemoveTask from the puller by the selected currency pair
func (p *RestPuller) RemoveTask(from string, to string) {
	t := p.Task(from, to)
	t.close()
	p.pullerMu.Lock()
	defer p.pullerMu.Unlock()
	delete(p.t, buildTaskName(from, to))
}
