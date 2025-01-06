package clients

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/session"
)

// RestPuller puller base struct
type RestPuller struct {
	t        Tasks
	s        *session.KeysStore
	dataPipe chan *Data
	client   RestClient
	pullerMu sync.RWMutex
}

// NewPuller init rest puller
func NewPuller(r RestClient, s *session.KeysStore, dataPipe chan *Data) RestApiPuller {
	return &RestPuller{
		t:        Tasks{},
		s:        s,
		dataPipe: dataPipe,
		client:   r,
	}
}

func (p *RestPuller) newTask(from string, to string, interval int64) *Task {
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
func (p *RestPuller) Task(from, to string) *Task {
	return p.task(buildTaskName(from, to))
}

func (p *RestPuller) task(name string) *Task {
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()
	return p.t[name]
}

func buildTaskName(from, to string) string {
	return strings.ToUpper(fmt.Sprintf("%s:%s", from, to))
}

// AddTask to collect data for the selected currency pair to the puller
func (p *RestPuller) AddTask(from string, to string, interval int64) *Task {
	t := p.newTask(from, to, interval)
	t.run(p.client, p.dataPipe)
	name := buildTaskName(from, to)
	p.pullerMu.Lock()
	p.t[name] = t
	p.pullerMu.Unlock()
	p.s.AppendTaskToSession(name, interval)
	return t
}

// RemoveTask from the puller by the selected currency pair
func (p *RestPuller) RemoveTask(from string, to string) {
	name := buildTaskName(from, to)
	t := p.task(name)
	t.close()
	p.pullerMu.Lock()
	defer p.pullerMu.Unlock()
	delete(p.t, name)
	p.s.RemoveTaskFromSession(name)
}

func (p *RestPuller) RestoreLastSession() (err error) {
	if p.s == nil {
		return
	}
	for k, v := range p.s.GetSession() {
		if pair := strings.Split(k, ":"); len(pair) == 2 {
			var i int64
			from, to := pair[0], pair[1]
			if i, err = strconv.ParseInt(v, 10, 64); err != nil {
				continue
			}
			p.AddTask(from, to, i)
		}
	}
	return
}

func (p *RestPuller) UpdateTask(t *Task, interval int64) *Task {
	atomic.StoreInt64(&t.Interval, interval)
	p.s.AppendTaskToSession(buildTaskName(t.From, t.To), interval)
	return t
}
