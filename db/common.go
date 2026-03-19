package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db/mysql"
	"github.com/streamdp/ccd/db/postgresql"
	"github.com/streamdp/ccd/domain"
)

// Database interface makes it possible to expand the list of data storages
type Database interface {
	Insert(ctx context.Context, data *domain.Data) (result sql.Result, err error)
	GetLast(ctx context.Context, from, to string) (result *domain.Data, err error)
	DataPipe() chan *domain.Data

	Close() error
}

func Connect(cfg *config.App) (any, error) {
	var (
		database any
		err      error
	)

	driverName, connectionString := getDataSource(cfg.DatabaseUrl)
	switch driverName {
	case postgresql.Postgres, "postgresql":
		database, err = postgresql.Connect(cfg.DatabaseUrl)
	case mysql.Mysql:
		fallthrough
	default:
		database, err = mysql.Connect(connectionString)
	}

	if err != nil {
		return nil, fmt.Errorf("database connection error: %w", err)
	}

	return database, nil
}

func Serve(ctx context.Context, d Database, l *log.Logger) {
	for data := range d.DataPipe() {
		if _, err := d.Insert(ctx, data); err != nil {
			l.Println(err)
		}
	}
}

func getDataSource(dataBaseUrl string) (string, string) {
	driverName := mysql.Mysql
	connectionString := dataBaseUrl

	if parameters := strings.Split(dataBaseUrl, "://"); len(parameters) == 2 {
		driverName, connectionString = parameters[0], parameters[1]
	}

	return driverName, connectionString
}
