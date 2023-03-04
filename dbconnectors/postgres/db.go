package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/lib/pq"

	"github.com/streamdp/ccd/clients"
)

const Postgres = "postgres"

// Db needed to add new methods for an instance *sql.Db
type Db struct {
	*sql.DB
}

// GetLast row with the most recent data for the selected currencies pair
func (db *Db) GetLast(from string, to string) (result *clients.Data, err error) {
	var displayDataRaw string
	result = clients.GetEmptyData(from, to)
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
		&result.Raw.Change24Hour,
		&result.Raw.Changepct24Hour,
		&result.Raw.Open24Hour,
		&result.Raw.Volume24Hour,
		&result.Raw.Low24Hour,
		&result.Raw.High24Hour,
		&result.Raw.Price,
		&result.Raw.Supply,
		&result.Raw.Mktcap,
		&result.Raw.Lastupdate,
		&displayDataRaw,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(displayDataRaw), &result.Display)
	result.Display.Lastupdate = time.Unix(result.Raw.Lastupdate, 0).String()
	return result, nil
}

// Insert clients.Data from the clients.DataPipe to the Db
func (db *Db) Insert(data *clients.Data) (result sql.Result, err error) {
	if data == nil || data.Raw == nil {
		return nil, errors.New("cant insert empty data")
	}
	d, err := json.Marshal(data.Display)
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
		&data.Raw.Change24Hour,
		&data.Raw.Changepct24Hour,
		&data.Raw.Open24Hour,
		&data.Raw.Volume24Hour,
		&data.Raw.Low24Hour,
		&data.Raw.High24Hour,
		&data.Raw.Price,
		&data.Raw.Supply,
		&data.Raw.Mktcap,
		&data.Raw.Lastupdate,
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
