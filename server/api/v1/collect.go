package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/domain"
	"github.com/streamdp/ccd/server/handlers"
)

type Puller interface {
	Task(from string, to string) *clients.Task
	AddTask(ctx context.Context, from string, to string, interval int64) *clients.Task
	RemoveTask(ctx context.Context, from string, to string)
	ListTasks() clients.Tasks
	UpdateTask(ctx context.Context, t *clients.Task, interval int64) *clients.Task
	RestoreLastSession(ctx context.Context) error
}

// CollectQuery structure for easily json serialization/validation/binding GET and POST query data
type CollectQuery struct {
	From     string `binding:"required,symbols" form:"fsym"     json:"fsym"`
	To       string `binding:"required,symbols" form:"tsym"     json:"tsym"`
	Interval int64  `form:"interval,default=60" json:"interval"`
}

func (c *CollectQuery) toUpper() {
	c.From = strings.ToUpper(c.From)
	c.To = strings.ToUpper(c.To)
}

// AddWorker that will collect data for the selected currency pair to the management service
func AddWorker(ctx context.Context, p Puller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := CollectQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.toUpper()

		if t := p.Task(q.From, q.To); t != nil {
			return domain.NewResult(
				http.StatusOK, "Data for this pair is already being collected", t,
			), nil
		}

		return domain.NewResult(
			http.StatusCreated,
			"Data collection started",
			p.AddTask(ctx, q.From, q.To, q.Interval),
		), nil
	}
}

// RemoveWorker from the management service and stop collecting data for the selected currencies pair
func RemoveWorker(ctx context.Context, p Puller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := CollectQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.toUpper()

		if p.Task(q.From, q.To) == nil {
			return domain.NewResult(
				http.StatusOK, "No data is collected for this pair", nil,
			), nil
		}
		p.RemoveTask(ctx, q.From, q.To)

		return domain.NewResult(http.StatusOK, "Task stopped successfully", nil), nil
	}
}

// PullingStatus return information about running pull tasks
func PullingStatus(p Puller, w clients.WsClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		var (
			tasks         clients.Tasks
			subscriptions domain.Subscriptions
		)
		if p != nil {
			tasks = p.ListTasks()
		}
		if w != nil {
			subscriptions = w.ListSubscriptions()
		}

		return domain.NewResult(
			http.StatusOK, "Information about running tasks", mergeTasks(tasks, subscriptions),
		), nil
	}
}

// UpdateWorker update pulling data interval for the selected worker by the currencies pair
func UpdateWorker(ctx context.Context, p Puller) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := CollectQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.toUpper()

		var t *clients.Task
		if t = p.Task(q.From, q.To); t == nil {
			return domain.NewResult(http.StatusOK, "No data is collected for this pair", t), nil
		}
		p.UpdateTask(ctx, t, q.Interval)

		return domain.NewResult(http.StatusOK, "Task updated successfully", t), nil
	}
}

func Subscribe(ctx context.Context, w clients.WsClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := CollectQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.toUpper()

		if err := w.Subscribe(ctx, q.From, q.To); err != nil {
			return &domain.Result{}, fmt.Errorf("subscribe error: %w", err)
		}

		return domain.NewResult(
			http.StatusCreated,
			"Subscribed successfully, data collection started",
			[]string{q.From, q.To},
		), nil
	}
}

func Unsubscribe(ctx context.Context, w clients.WsClient) handlers.HandlerFuncResError {
	return func(c *gin.Context) (*domain.Result, error) {
		q := CollectQuery{}
		if err := c.Bind(&q); err != nil {
			return &domain.Result{}, fmt.Errorf("%w: %w", handlers.ErrBindQuery, err)
		}
		q.toUpper()

		if err := w.Unsubscribe(ctx, q.From, q.To); err != nil {
			return &domain.Result{}, fmt.Errorf("unsubscribe error: %w", err)
		}

		return domain.NewResult(
			http.StatusOK,
			"Unsubscribed successfully, data collection stopped",
			[]string{q.From, q.To},
		), nil
	}
}

func mergeTasks(tasks clients.Tasks, subscriptions domain.Subscriptions) any {
	if len(tasks) == 0 && len(subscriptions) == 0 {
		return nil
	}

	list := map[string]map[string]interface{}{}

	if len(tasks) != 0 {
		for _, v := range tasks {
			if list[v.From] != nil {
				list[v.From][v.To] = v

				continue
			}
			list[v.From] = make(map[string]interface{})
			list[v.From][v.To] = v
		}
	}

	if len(subscriptions) != 0 {
		for _, v := range subscriptions {
			if list[v.From] != nil {
				list[v.From][v.To] = v

				continue
			}
			list[v.From] = make(map[string]interface{})
			list[v.From][v.To] = v
		}
	}

	return list
}
