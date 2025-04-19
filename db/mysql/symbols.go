package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/streamdp/ccd/domain"
)

var (
	errEmptySymbol  = errors.New("empty symbol")
	errExecuteQuery = errors.New("failed to execute query")
	errCopyResult   = errors.New("failed to copy result")
	errParseResults = errors.New("failed to parse results")
)

func (d *Db) AddSymbol(s string, u string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	result, err := d.Exec(
		`insert into symbols (symbol,unicode) values (?,?);`, strings.ToUpper(s), strings.ToUpper(u),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) UpdateSymbol(s, u string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	result, err := d.Exec(`update symbols set unicode=? where symbol=?;`, strings.ToUpper(u), strings.ToUpper(s))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) RemoveSymbol(s string) (sql.Result, error) {
	if s == "" {
		return nil, errEmptySymbol
	}

	result, err := d.Exec(`delete from symbols where symbol=?;`, strings.ToUpper(s))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) Symbols() ([]*domain.Symbol, error) {
	rows, errQuery := d.Query(`select symbol, unicode from symbols`)
	if errQuery != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, errQuery)
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
			return nil, fmt.Errorf("%w: %w", errCopyResult, err)
		}
		r, _ := utf8.DecodeRune(u)
		symbols = append(symbols, &domain.Symbol{
			Symbol:  s,
			Unicode: r,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errParseResults, err)
	}

	return symbols, nil
}
