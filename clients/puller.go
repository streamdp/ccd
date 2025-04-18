package clients

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db"
	"github.com/streamdp/ccd/domain"
)

// RestApiPuller interface makes it possible to expand the list of rest api pullers
type RestApiPuller interface {
	Task(from string, to string) *Task
	AddTask(from string, to string, interval int64) *Task
	RemoveTask(from string, to string)
	ListTasks() Tasks
	UpdateTask(t *Task, interval int64) *Task
	RestoreLastSession() error
}

// restPuller puller base struct
type restPuller struct {
	t        Tasks
	l        *log.Logger
	s        db.Session
	dataPipe chan *domain.Data
	client   RestClient
	pullerMu sync.RWMutex
}

// NewPuller init rest puller
func NewPuller(r RestClient, l *log.Logger, s db.Session, dataPipe chan *domain.Data) *restPuller {
	return &restPuller{
		t:        Tasks{},
		l:        l,
		s:        s,
		dataPipe: dataPipe,
		client:   r,
	}
}

// ListTasks return all tasks
func (p *restPuller) ListTasks() Tasks {
	var t = make(Tasks, len(p.t))
	p.pullerMu.RLock()
	for k, v := range p.t {
		t[k] = v
	}
	p.pullerMu.RUnlock()

	return t
}

// Task return task with selected currencies pair, if possible
func (p *restPuller) Task(from, to string) *Task {
	return p.task(buildTaskName(from, to))
}

// AddTask to collect data for the selected currency pair to the puller
func (p *restPuller) AddTask(from string, to string, interval int64) *Task {
	t := p.newTask(from, to, interval)
	t.run(p.client, p.l, p.dataPipe)
	name := buildTaskName(from, to)
	p.pullerMu.Lock()
	p.t[name] = t
	p.pullerMu.Unlock()
	if err := p.s.AddTask(name, interval); err != nil {
		p.l.Println(err)
	}

	return t
}

// RemoveTask from the puller by the selected currency pair
func (p *restPuller) RemoveTask(from string, to string) {
	name := buildTaskName(from, to)
	t := p.task(name)
	t.close()
	p.pullerMu.Lock()
	delete(p.t, name)
	p.pullerMu.Unlock()
	if err := p.s.RemoveTask(name); err != nil {
		p.l.Print(err)
	}
}

// RestoreLastSession get the last session from the session store and restore it
func (p *restPuller) RestoreLastSession() error {
	if p.s == nil {
		return nil
	}
	ses, err := p.s.GetSession()
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	for k, v := range ses {
		if pair := strings.Split(k, ":"); len(pair) == 2 {
			from, to := pair[0], pair[1]
			p.AddTask(from, to, v)
		}
	}

	return nil
}

func (p *restPuller) UpdateTask(t *Task, interval int64) *Task {
	atomic.StoreInt64(&t.Interval, interval)
	if err := p.s.UpdateTask(buildTaskName(t.From, t.To), interval); err != nil {
		p.l.Println(err)
	}

	return t
}

func buildTaskName(from, to string) string {
	return strings.ToUpper(fmt.Sprintf("%s:%s", from, to))
}

func (p *restPuller) newTask(from string, to string, interval int64) *Task {
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

func (p *restPuller) task(name string) *Task {
	p.pullerMu.RLock()
	defer p.pullerMu.RUnlock()

	return p.t[name]
}
