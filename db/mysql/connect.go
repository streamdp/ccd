package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/streamdp/ccd/domain"
)

const Mysql = "mysql"

// Db needed to add new methods for an instance *sql.Db
type Db struct {
	*sql.DB
	pipe chan *domain.Data
}

func (d *Db) DataPipe() chan *domain.Data {
	return d.pipe
}

// Connect after prepare to the Db
func Connect(dataSource string) (db *Db, err error) {
	sqlDb := &sql.DB{}
	if sqlDb, err = sql.Open(Mysql, dataSource); err != nil {
		return
	}
	return &Db{
		DB:   sqlDb,
		pipe: make(chan *domain.Data, 1000),
	}, nil
}

// Close Db connection
func (d *Db) Close() (err error) {
	defer close(d.pipe)
	if d.DB == nil {
		return
	}
	return d.DB.Close()
}
