package clients

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/domain"
)

type restPuller struct {
	tasks       Tasks
	l           *log.Logger
	sessionRepo SessionRepo
	dataPipe    chan *domain.Data
	client      RestClient
	pullerMu    sync.RWMutex
}

func NewPuller(r RestClient, l *log.Logger, sessionRepo SessionRepo, dataPipe chan *domain.Data) *restPuller {
	return &restPuller{
		tasks:       Tasks{},
		l:           l,
		sessionRepo: sessionRepo,
		dataPipe:    dataPipe,
		client:      r,
	}
}

// ListTasks return all tasks
func (p *restPuller) ListTasks() Tasks {
	var t = make(Tasks, len(p.tasks))

	p.pullerMu.RLock()
	for k, v := range p.tasks {
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
func (p *restPuller) AddTask(ctx context.Context, from string, to string, interval int64) *Task {
	name := buildTaskName(from, to)

	t := p.newTask(from, to, interval)
	t.run(p.client, p.l, p.dataPipe)

	p.pullerMu.Lock()
	p.tasks[name] = t
	p.pullerMu.Unlock()

	if err := p.sessionRepo.AddTask(ctx, name, interval); err != nil {
		p.l.Println(err)
	}

	return t
}

// RemoveTask from the puller by the selected currency pair
func (p *restPuller) RemoveTask(ctx context.Context, from string, to string) {
	name := buildTaskName(from, to)

	t := p.task(name)
	t.close()

	p.pullerMu.Lock()
	delete(p.tasks, name)
	p.pullerMu.Unlock()

	if err := p.sessionRepo.RemoveTask(ctx, name); err != nil {
		p.l.Print(err)
	}
}

// RestoreLastSession get the last session from the session store and restore it
func (p *restPuller) RestoreLastSession(ctx context.Context) error {
	if p.sessionRepo == nil {
		return nil
	}
	ses, err := p.sessionRepo.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	for k, v := range ses {
		if pair := strings.Split(k, ":"); len(pair) == 2 {
			from, to := pair[0], pair[1]
			p.AddTask(ctx, from, to, v)
		}
	}

	return nil
}

func (p *restPuller) UpdateTask(ctx context.Context, t *Task, interval int64) *Task {
	atomic.StoreInt64(&t.Interval, interval)
	if err := p.sessionRepo.UpdateTask(ctx, buildTaskName(t.From, t.To), interval); err != nil {
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

	return p.tasks[name]
}
