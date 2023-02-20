package dbconnectors

import (
	"database/sql"
	"encoding/json"
	"errors"
	// this is a mysql-driver, it is needed for database/sql to work correctly
	_ "github.com/go-sql-driver/mysql"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/handlers"
	"time"
)

// Db needed to add new methods for an instance *sql.Db
type Db struct {
	*sql.DB
}

// GetLast row with the most recent data for the selected currencies pair
func (db *Db) GetLast(from string, to string) (result *dataproviders.Data, err error) {
	var displayDataRaw string
	result = dataproviders.GetEmptyData(from, to)
	r := result.Raw[from][to]
	row := db.QueryRow("select change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate, displaydataraw from cryptocompare.data where fromSym=(select _id from fsym where symbol=?) and toSym=(select _id from tsym where symbol=?) ORDER BY lastupdate DESC limit 1;", from, to)
	err = row.Scan(&r.Change24Hour, &r.Changepct24Hour, &r.Open24Hour, &r.Volume24Hour, &r.Low24Hour, &r.High24Hour, &r.Price, &r.Supply, &r.Mktcap, &r.Lastupdate, &displayDataRaw)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(displayDataRaw), &result.Display)
	result.Display[from][to].Lastupdate = time.Unix(result.Raw[from][to].Lastupdate, 0).String()
	return result, nil
}

// Insert dataproviders.Data from the dataproviders.DataPipe to the Db
func (db *Db) Insert(data *dataproviders.DataPipe) (result sql.Result, err error) {
	if data.Data == nil || data.Data.Raw[data.From] == nil {
		return nil, errors.New("cant insert empty data")
	}
	r := data.Data.Raw[data.From][data.To]
	d, err := json.Marshal(data.Data.Display)
	if err != nil {
		return nil, err
	}
	result, err = db.Exec("insert into data (fromSym, toSym, change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate, displaydataraw) values ((SELECT _id FROM fsym WHERE symbol=?),(SELECT _id FROM tsym WHERE symbol=?),?,?,?,?,?,?,?,?,?,?,?)", data.From, data.To, &r.Change24Hour, &r.Changepct24Hour, &r.Open24Hour, &r.Volume24Hour, &r.Low24Hour, &r.High24Hour, &r.Price, &r.Supply, &r.Mktcap, &r.Lastupdate, string(d))
	if err != nil {
		return nil, err
	}
	return
}

// Close Db connection
func (db *Db) Close() (err error) {
	return db.Close()
}

// Connect after prepare to the Db
func Connect() (db *Db, err error) {
	dataSource := config.GetEnv("CCDC_DATASOURCE")
	if dataSource == "" {
		return nil, errors.New("please set OS environment \"CCDC_DATASOURCE\" with database connection string")
	}
	sqlDb := &sql.DB{}
	if sqlDb, err = sql.Open("mysql", dataSource); err != nil {
		return
	}
	return &Db{sqlDb}, nil
}

// ServePipe and if we get dataproviders.DataPipe then save them to the Db
func (db *Db) ServePipe(pipe chan *dataproviders.DataPipe) {
	go func() {
		for data := range pipe {
			if _, err := db.Insert(data); err != nil {
				handlers.SystemHandler(err)
			}
		}
	}()
}
