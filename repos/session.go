package repos

import (
	"database/sql"
	"fmt"
)

type TaskStore interface {
	AddTask(n string, i int64) (result sql.Result, err error)
	UpdateTask(n string, i int64) (result sql.Result, err error)
	RemoveTask(n string) (result sql.Result, err error)
	GetSession() (tasks map[string]int64, err error)
}

type sessionRepo struct {
	r TaskStore
}

func NewSessionRepo(r TaskStore) (*sessionRepo, error) {
	return &sessionRepo{
		r: r,
	}, nil
}

func (sr *sessionRepo) UpdateTask(n string, i int64) error {
	if _, err := sr.r.UpdateTask(n, i); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) GetSession() (map[string]int64, error) {
	session, err := sr.r.GetSession()
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (sr *sessionRepo) AddTask(n string, i int64) error {
	_, err := sr.r.AddTask(n, i)
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) RemoveTask(n string) error {
	_, err := sr.r.RemoveTask(n)
	if err != nil {
		return fmt.Errorf("failed to remove task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) Close() error {
	return nil
}
