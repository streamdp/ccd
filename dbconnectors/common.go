package dbconnectors

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/streamdp/ccd/clients"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dbconnectors/mysql"
	"github.com/streamdp/ccd/dbconnectors/postgres"
	"github.com/streamdp/ccd/handlers"
)

// DbReadWrite interface makes it possible to expand the list of data storages
type DbReadWrite interface {
	Insert(data *clients.Data) (result sql.Result, err error)
	GetLast(from string, to string) (result *clients.Data, err error)
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
		return postgres.Connect(dataSource)
	case mysql.Mysql:
		fallthrough
	default:
		return mysql.Connect(connectionString)
	}
}

// ServePipe and if we get clients.DataPipe then save them to the Db
func ServePipe(db DbReadWrite, pipe chan *clients.Data) {
	go func() {
		for data := range pipe {
			if _, err := db.Insert(data); err != nil {
				handlers.SystemHandler(err)
			}
		}
	}()
}
