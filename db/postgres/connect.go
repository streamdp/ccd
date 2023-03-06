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
	pipe chan *clients.Data
}

// GetLast row with the most recent data for the selected currencies pair
func (d *Db) GetLast(from string, to string) (result *clients.Data, err error) {
	var displayDataRaw string
	result = clients.EmptyData(from, to)
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
	err = d.QueryRow(query, from, to).Scan(
		&result.Raw.Change24Hour,
		&result.Raw.Changepct24Hour,
		&result.Raw.Open24Hour,
		&result.Raw.Volume24Hour,
		&result.Raw.Low24Hour,
		&result.Raw.High24Hour,
		&result.Raw.Price,
		&result.Raw.Supply,
		&result.Raw.MktCap,
		&result.Raw.LastUpdate,
		&displayDataRaw,
	)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(displayDataRaw), &result.Display)
	result.Display.LastUpdate = time.Unix(result.Raw.LastUpdate, 0).String()
	return result, nil
}

// Insert clients.Data from the clients.DataPipe to the Db
func (d *Db) Insert(data *clients.Data) (result sql.Result, err error) {
	if data == nil || data.Raw == nil {
		return nil, errors.New("cant insert empty data")
	}
	dsp, err := json.Marshal(data.Display)
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
	return d.Exec(
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
		&data.Raw.MktCap,
		&data.Raw.LastUpdate,
		string(dsp),
	)
}

func (d *Db) DataPipe() chan *clients.Data {
	return d.pipe
}

// Connect after prepare to the Db
func Connect(dataSource string) (d *Db, err error) {
	sqlDb := &sql.DB{}
	if sqlDb, err = sql.Open(Postgres, dataSource); err != nil {
		return
	}
	return &Db{
		DB:   sqlDb,
		pipe: make(chan *clients.Data, 1000),
	}, nil
}

// Close Db connection
func (d *Db) Close() (err error) {
	return d.Close()
}
