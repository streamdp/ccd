package postgresql

import (
	"database/sql"
	"errors"
	"strings"
)

var (
	errEmptyTaskName = errors.New("empty task name")
	errEmptySymbol   = errors.New("empty symbol")
)

func (d *Db) AddTask(n string, i int64) (sql.Result, error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	return d.Exec(
		`insert into session (task_name,interval) values ($1,$2) on conflict do nothing;`, strings.ToUpper(n), i,
	)
}

func (d *Db) UpdateTask(n string, i int64) (sql.Result, error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	return d.Exec(`update session set interval=$2 where task_name=$1;`, strings.ToUpper(n), i)
}

func (d *Db) RemoveTask(n string) (sql.Result, error) {
	if n == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`delete from session where task_name=$1;`, strings.ToUpper(n))
}

func (d *Db) GetSession() (map[string]int64, error) {
	rows, err := d.Query(`select task_name,"interval" from session`)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	tasks := make(map[string]int64)
	for rows.Next() {
		var (
			n string
			i int64
		)
		if err = rows.Scan(&n, &i); err != nil {
			return nil, err
		}
		tasks[n] = i
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return tasks, nil
}
