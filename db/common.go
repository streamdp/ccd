package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db/mysql"
	"github.com/streamdp/ccd/db/postgres"
	"github.com/streamdp/ccd/handlers"
)

// Database interface makes it possible to expand the list of data storages
type Database interface {
	Insert(data *clients.Data) (result sql.Result, err error)
	GetLast(from string, to string) (result *clients.Data, err error)
	DataPipe() chan *clients.Data
}

func Connect() (d Database, err error) {
	var (
		driverName       = mysql.Mysql
		dataSource       = config.GetEnv("CCDC_DATASOURCE")
		connectionString string
	)
	if dataSource == "" {
		return nil, errors.New("please set OS environment \"CCDC_DATASOURCE\" with database connection string")
	}
	connectionParameters := strings.Split(dataSource, "://")
	if len(connectionParameters) == 2 {
		driverName, connectionString = connectionParameters[0], connectionParameters[1]
	}
	switch driverName {
	case postgres.Postgres:
		d, err = postgres.Connect(dataSource)
	case mysql.Mysql:
		fallthrough
	default:
		d, err = mysql.Connect(connectionString)
	}
	if err == nil {
		serve(d)
	}
	return
}

func serve(d Database) {
	go func() {
		for data := range d.DataPipe() {
			if _, err := d.Insert(data); err != nil {
				handlers.SystemHandler(err)
			}
		}
	}()
}
