package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/streamdp/ccd/domain"
)

var errEmptyData = errors.New("cant insert empty data")

// GetLast row with the most recent data for the selected currencies pair
func (d *Db) GetLast(from string, to string) (*domain.Data, error) {
	result := &domain.Data{
		FromSymbol: from,
		ToSymbol:   to,
	}
	query := `
		select
		    _id,
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
		from data 
		where fromSym=(select _id from symbols where symbol=?) 
		  and toSym=(select _id from symbols where symbol=?) 
		ORDER BY lastupdate DESC limit 1;
`
	if err := d.QueryRow(query, from, to).Scan(
		&result.Id,
		&result.Change24Hour,
		&result.ChangePct24Hour,
		&result.Open24Hour,
		&result.Volume24Hour,
		&result.Low24Hour,
		&result.High24Hour,
		&result.Price,
		&result.Supply,
		&result.MktCap,
		&result.LastUpdate,
		&result.DisplayDataRaw,
	); err != nil {
		return nil, fmt.Errorf("%w: %w", errCopyResult, err)
	}

	return result, nil
}

// Insert clients.Data from the clients.DataPipe to the Db
func (d *Db) Insert(data *domain.Data) (sql.Result, error) {
	if data == nil {
		return nil, errEmptyData
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
		        (SELECT _id FROM symbols WHERE symbol=?),
		        (SELECT _id FROM symbols WHERE symbol=?),
		        ?,?,?,?,?,?,?,?,?,?,?
		)
`
	result, err := d.Exec(
		query,
		&data.FromSymbol,
		&data.ToSymbol,
		&data.Change24Hour,
		&data.ChangePct24Hour,
		&data.Open24Hour,
		&data.Volume24Hour,
		&data.Low24Hour,
		&data.High24Hour,
		&data.Price,
		&data.Supply,
		&data.MktCap,
		&data.LastUpdate,
		&data.DisplayDataRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}
