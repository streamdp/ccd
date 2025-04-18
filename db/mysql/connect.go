package mysql

import (
	"database/sql"
	"fmt"

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
func Connect(dataSource string) (*Db, error) {
	sqlDb, err := sql.Open(Mysql, dataSource)
	if err != nil {
		return nil, err
	}

	return &Db{
		DB:   sqlDb,
		pipe: make(chan *domain.Data, 1000),
	}, nil
}

// Close Db connection
func (d *Db) Close() error {
	defer close(d.pipe)

	if d.DB == nil {
		return nil
	}

	if err := d.DB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	return nil
}
