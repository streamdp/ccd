package postgresql

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/streamdp/ccd/domain"
)

const Postgres = "postgres"

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
	sqlDb, err := sql.Open(Postgres, dataSource)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
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
