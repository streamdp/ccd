package repos

import (
	"fmt"

	"github.com/streamdp/ccd/db"
)

type sessionRepo struct {
	db db.Database
}

func NewSessionRepo(db db.Database) (*sessionRepo, error) {
	return &sessionRepo{
		db: db,
	}, nil
}

func (sr *sessionRepo) UpdateTask(n string, i int64) error {
	if _, err := sr.db.UpdateTask(n, i); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (sr *sessionRepo) GetSession() (map[string]int64, error) {
	return sr.db.GetSession()
}

func (sr *sessionRepo) AddTask(n string, i int64) (err error) {
	_, err = sr.db.AddTask(n, i)

	return
}

func (sr *sessionRepo) RemoveTask(n string) (err error) {
	_, err = sr.db.RemoveTask(n)

	return
}

func (sr *sessionRepo) Close() error {
	return nil
}
