package postgresql

import (
	"database/sql"
	"strings"
	"unicode/utf8"

	"github.com/streamdp/ccd/domain"
)

func (d *Db) AddSymbol(s, u string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`insert into symbols (symbol,unicode) values ($1,$2);`, strings.ToUpper(s), strings.ToUpper(u))
}

func (d *Db) UpdateSymbol(s, u string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`update symbols set unicode=$2 where symbol=$1;`, strings.ToUpper(s), strings.ToUpper(u))
}

func (d *Db) RemoveSymbol(s string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`delete from symbols where symbol=$1;`, strings.ToUpper(s))
}

func (d *Db) Symbols() ([]*domain.Symbol, error) {
	rows, errQuery := d.Query(`select symbol, unicode from symbols`)
	if errQuery != nil {
		return nil, errQuery
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var symbols []*domain.Symbol
	for rows.Next() {
		var (
			s string
			u []byte
		)
		if err := rows.Scan(&s, &u); err != nil {
			return nil, err
		}
		r, _ := utf8.DecodeRune(u)
		symbols = append(symbols, &domain.Symbol{
			Symbol:  s,
			Unicode: r,
		})
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return symbols, nil
}
