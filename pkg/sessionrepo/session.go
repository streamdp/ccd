package sessionrepo

import (
	"context"
	"database/sql"
	"fmt"
)

type SessionStore interface {
	AddTask(ctx context.Context, n string, i int64) (result sql.Result, err error)
	UpdateTask(ctx context.Context, n string, i int64) (result sql.Result, err error)
	RemoveTask(ctx context.Context, n string) (result sql.Result, err error)
	GetSession(ctx context.Context) (tasks map[string]int64, err error)
}

type sessionRepo struct {
	r SessionStore
}

func New(r SessionStore) (*sessionRepo, error) {
	return &sessionRepo{
		r: r,
	}, nil
}

func (sr *sessionRepo) UpdateTask(ctx context.Context, n string, i int64) error {
	if _, err := sr.r.UpdateTask(ctx, n, i); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) GetSession(ctx context.Context) (map[string]int64, error) {
	session, err := sr.r.GetSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (sr *sessionRepo) AddTask(ctx context.Context, n string, i int64) error {
	if _, err := sr.r.AddTask(ctx, n, i); err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) RemoveTask(ctx context.Context, n string) error {
	if _, err := sr.r.RemoveTask(ctx, n); err != nil {
		return fmt.Errorf("failed to remove task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) Close() error {
	return nil
}
