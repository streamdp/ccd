package mysql

import (
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/streamdp/ccd/domain"
)

var errEmptySymbol = errors.New("cant insert empty symbol")

func (d *Db) AddSymbol(s string, u string) (result sql.Result, err error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`insert into symbols (symbol,unicode) values (?,?);`, strings.ToUpper(s), strings.ToUpper(u))
}

func (d *Db) UpdateSymbol(s, u string) (result sql.Result, err error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`update symbols set unicode=? where symbol=?;`, strings.ToUpper(u), strings.ToUpper(s))
}

func (d *Db) RemoveSymbol(s string) (result sql.Result, err error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`delete from symbols where symbol=?;`, strings.ToUpper(s))
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return symbols, nil
}
