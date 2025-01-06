package repos

import (
	"github.com/streamdp/ccd/db"
)

type SessionRepo struct {
	db db.Database
}

func NewSessionRepo(db db.Database) (db.Session, error) {
	return &SessionRepo{
		db: db,
	}, nil
}

func (sr *SessionRepo) UpdateTask(n string, i int64) (err error) {
	_, err = sr.db.UpdateTask(n, i)
	return
}

func (sr *SessionRepo) GetSession() (map[string]int64, error) {
	return sr.db.GetSession()
}

func (sr *SessionRepo) AddTask(n string, i int64) (err error) {
	_, err = sr.db.AddTask(n, i)
	return
}

func (sr *SessionRepo) RemoveTask(n string) (err error) {
	_, err = sr.db.RemoveTask(n)
	return
}
