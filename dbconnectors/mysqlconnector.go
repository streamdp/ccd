package dbconnectors

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streamdp/ccd/config"
	"github.com/streamdp/ccd/dataproviders"
	"github.com/streamdp/ccd/handlers"
	"time"
)

type Db struct {
	instance *sql.DB
}

func (db *Db) GetLast(from string, to string) (result *dataproviders.Data, err error) {
	var displayDataRaw string
	result = dataproviders.GetEmptyData(from, to)
	r := result.Raw[from][to]
	row := db.instance.QueryRow("select change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate, displaydataraw from cryptocompare.data where fromSym=(select _id from fsym where symbol=?) and toSym=(select _id from tsym where symbol=?) ORDER BY lastupdate DESC limit 1;", from, to)
	err = row.Scan(&r.Change24Hour, &r.Changepct24Hour, &r.Open24Hour, &r.Volume24Hour, &r.Low24Hour, &r.High24Hour, &r.Price, &r.Supply, &r.Mktcap, &r.Lastupdate, &displayDataRaw)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(displayDataRaw), &result.Display)
	result.Display[from][to].Lastupdate = time.Unix(result.Raw[from][to].Lastupdate, 0).String()
	return result, nil
}

func (db *Db) Insert(data *dataproviders.DataPipe) (result sql.Result, err error) {
	if data.Data == nil || data.Data.Raw[data.From] == nil {
		return nil, errors.New("cant insert empty data")
	}
	r := data.Data.Raw[data.From][data.To]
	d, err := json.Marshal(data.Data.Display)
	if err != nil {
		return nil, err
	}
	result, err = db.instance.Exec("insert into data (fromSym, toSym, change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate, displaydataraw) values ((SELECT _id FROM fsym WHERE symbol=?),(SELECT _id FROM tsym WHERE symbol=?),?,?,?,?,?,?,?,?,?,?,?)", data.From, data.To, &r.Change24Hour, &r.Changepct24Hour, &r.Open24Hour, &r.Volume24Hour, &r.Low24Hour, &r.High24Hour, &r.Price, &r.Supply, &r.Mktcap, &r.Lastupdate, string(d))
	if err != nil {
		return nil, err
	}
	return
}

func (db *Db) Connect(dataSource string) (err error) {
	if db.instance, err = sql.Open("mysql", dataSource); err != nil {
		return err
	}
	return nil
}

func (db *Db) Close() (err error) {
	return db.instance.Close()
}

func Connect() (db *Db, err error) {
	db = &Db{}
	dataSource := config.GetEnv("CCDC_DATASOURCE")
	if dataSource == "" {
		return nil, errors.New("please set OS environment \"CCDC_DATASOURCE\" with database connection string")
	}
	if err = db.Connect(dataSource); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Db) ServePipe(pipe chan *dataproviders.DataPipe) {
	for {
		if data := <-pipe; true {
			if _, err := db.Insert(data); err != nil {
				handlers.SystemHandler(err)
			}
		}
	}
}
