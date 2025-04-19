package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/db/mysql"
	"github.com/streamdp/ccd/db/postgresql"
	"github.com/streamdp/ccd/domain"
)

// Session interface makes it possible to expand the list of session storages
type Session interface {
	AddTask(n string, i int64) (err error)
	UpdateTask(n string, i int64) (err error)
	RemoveTask(n string) (err error)
	GetSession() (map[string]int64, error)

	Close() error
}

// Database interface makes it possible to expand the list of data storages
type Database interface {
	Insert(data *domain.Data) (result sql.Result, err error)
	GetLast(from string, to string) (result *domain.Data, err error)
	DataPipe() chan *domain.Data

	Close() error
}

var errDatabaseUrl = errors.New("\"CCDC_DATABASEURL\" os environment variable is not set")

func Connect() (any, error) {
	dataBaseUrl, err := getDatabaseUrl()
	if err != nil {
		return nil, fmt.Errorf("failed to get database url: %w", err)
	}

	driverName, connectionString := getDataSource(dataBaseUrl)

	var database any
	switch driverName {
	case postgresql.Postgres, "postgresql":
		database, err = postgresql.Connect(dataBaseUrl)
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

func Serve(d Database, l *log.Logger) {
	for data := range d.DataPipe() {
		if _, err := d.Insert(data); err != nil {
			l.Println(err)
		}
	}
}

func getDatabaseUrl() (string, error) {
	var dataBaseUrl = config.GetEnv("CCDC_DATABASEURL")
	if dataBaseUrl == "" {
		return "", errDatabaseUrl
	}

	return dataBaseUrl, nil
}

func getDataSource(dataBaseUrl string) (string, string) {
	driverName := mysql.Mysql
	connectionString := dataBaseUrl

	if parameters := strings.Split(dataBaseUrl, "://"); len(parameters) == 2 {
		driverName, connectionString = parameters[0], parameters[1]
	}

	return driverName, connectionString
}
