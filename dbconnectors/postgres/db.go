package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/lib/pq"

	"github.com/streamdp/ccd/dataproviders"
)

const Postgres = "postgres"

// Db needed to add new methods for an instance *sql.Db
type Db struct {
	*sql.DB
}

// GetLast row with the most recent data for the selected currencies pair
func (db *Db) GetLast(from string, to string) (result *dataproviders.Data, err error) {
	var displayDataRaw string
	result = dataproviders.GetEmptyData(from, to)
	r := result.Raw[from][to]
	query := `
		select change24hour,
		       changepct24hour,
		       open24hour,
		       volume24hour,
		       low24hour,
		       high24hour, 
		       price, 
		       supply,
		       mktcap, 
		       lastupdate,
		       displaydataraw
		from data 
		where fromSym=(select _id from fsym where symbol=$1)
		  and toSym=(select _id from tsym where symbol=$2)
		ORDER BY lastupdate DESC limit 1;
`
	err = db.QueryRow(query, from, to).Scan(
		&r.Change24Hour,
		&r.Changepct24Hour,
		&r.Open24Hour,
		&r.Volume24Hour,
		&r.Low24Hour,
		&r.High24Hour,
		&r.Price,
		&r.Supply,
		&r.Mktcap,
		&r.Lastupdate,
		&displayDataRaw,
	)
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
	query := `insert into data (
                  fromSym,
                  toSym,
                  change24hour,
                  changepct24hour,
                  open24hour,
                  volume24hour,
                  low24hour,
                  high24hour, 
                  price,
                  supply, 
                  mktcap, 
                  lastupdate,
                  displaydataraw
        ) 
		values (
		        (SELECT _id FROM fsym WHERE symbol=$1),
		        (SELECT _id FROM tsym WHERE symbol=$2),
		        $3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13
		)
`
	return db.Exec(
		query,
		data.From,
		data.To,
		&r.Change24Hour,
		&r.Changepct24Hour,
		&r.Open24Hour,
		&r.Volume24Hour,
		&r.Low24Hour,
		&r.High24Hour,
		&r.Price,
		&r.Supply,
		&r.Mktcap,
		&r.Lastupdate,
		string(d),
	)
}

// Connect after prepare to the Db
func Connect(dataSource string) (db *Db, err error) {
	sqlDb := &sql.DB{}
	if sqlDb, err = sql.Open(Postgres, dataSource); err != nil {
		return
	}
	return &Db{sqlDb}, nil
}

// Close Db connection
func (db *Db) Close() (err error) {
	return db.Close()
}
