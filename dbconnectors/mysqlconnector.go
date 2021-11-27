package dbconnectors

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/streamdp/ccdatacollector/dataproviders"
	"github.com/streamdp/ccdatacollector/utility"
)

type Db struct {
	instance *sql.DB
}

func (d *Db) GetLast(from string, to string) (result *dataproviders.Data, err error) {
	r := &dataproviders.Response{}
	result = dataproviders.GetEmptyData(from, to)
	row := d.instance.QueryRow("select change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate from cryptocompare.data where fromSym=(select _id from fsym where symbol=?) and toSym=(select _id from tsym where symbol=?) ORDER BY lastupdate DESC limit 1;", from, to)
	err = row.Scan(&r.Change24Hour, &r.Changepct24Hour, &r.Open24Hour, &r.Volume24Hour, &r.Low24Hour, &r.High24Hour, &r.Price, &r.Supply, &r.Mktcap, &r.Lastupdate)
	result.Raw[from][to] = *r
	if err != nil {
		return &dataproviders.Data{}, errors.Wrapf(err, "can't match data between row and dest")
	}
	return result, nil
}

func (d *Db) Insert(data *dataproviders.DataPipe) (result sql.Result, err error) {
	r := data.Data.Raw[data.From][data.To]
	result, err = d.instance.Exec("insert into data (fromSym, toSym, change24hour, changepct24hour, open24hour, volume24hour, low24hour, high24hour, price, supply, mktcap, lastupdate) values ((SELECT _id FROM fsym WHERE symbol=?),(SELECT _id FROM tsym WHERE symbol=?),?,?,?,?,?,?,?,?,?,?)", data.From, data.To, r.Change24Hour, r.Changepct24Hour, r.Open24Hour, r.Volume24Hour, r.Low24Hour, r.High24Hour, r.Price, r.Supply, r.Mktcap, r.Lastupdate)
	if err != nil {
		return nil, errors.Wrapf(err, "can't insert data")
	}
	return
}

func (d *Db) Connect(dataSource string) (err error) {
	if d.instance, err = sql.Open("mysql", dataSource); err != nil {
		return errors.Wrapf(err, "can't connect to database")
	}
	return nil
}

func (d *Db) Close() (err error) {
	return d.instance.Close()
}

func ServePipe(pipe chan *dataproviders.DataPipe) {
	var db Db
	dataSource := utility.GetEnv("CCDC_DATASOURCE")
	if dataSource == "" {
		utility.HandleError(errors.New("please set OS environment \"CCDC_DATASOURCE\" with database connection string"))
		return
	}
	if err := db.Connect(dataSource); err != nil {
		utility.HandleError(err)
		return
	}
	defer func(d *Db) {
		if err := d.Close(); err != nil {
			utility.HandleError(err)
		}
	}(&db)
	for {
		if data := <-pipe; true {
			if _, err := db.Insert(data); err != nil {
				utility.HandleError(err)
			}
		}
	}
}
