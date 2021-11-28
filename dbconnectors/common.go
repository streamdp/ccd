package dbconnectors

import (
	"database/sql"
	"github.com/streamdp/ccdatacollector/dataproviders"
)

type DbReadWrite interface {
	Insert(data *dataproviders.DataPipe) (result sql.Result, err error)
	GetLast(from string, to string) (result *dataproviders.Data, err error)
}
