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

// DbReadWrite interface makes it possible to expand the list of data storages
type DbReadWrite interface {
	Insert(data *clients.Data) (result sql.Result, err error)
	GetLast(from string, to string) (result *clients.Data, err error)
	DataPipe() chan *clients.Data
}

func Connect() (db DbReadWrite, err error) {
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
		db, err = postgres.Connect(dataSource)
	case mysql.Mysql:
		fallthrough
	default:
		db, err = mysql.Connect(connectionString)
	}
	if err == nil {
		serve(db)
	}
	return
}

func serve(db DbReadWrite) {
	go func() {
		for data := range db.DataPipe() {
			if _, err := db.Insert(data); err != nil {
				handlers.SystemHandler(err)
			}
		}
	}()
}
