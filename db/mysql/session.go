package mysql

import (
	"database/sql"
	"errors"
	"strings"
)

var errEmptyTaskName = errors.New("empty task name")

func (d *Db) AddTask(n string, i int64) (result sql.Result, err error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	return d.Exec(
		`insert ignore into session (task_name,"interval") values (?,?);`, strings.ToUpper(n), i,
	)
}

func (d *Db) UpdateTask(n string, i int64) (result sql.Result, err error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	return d.Exec(`update session set "interval"=? where task_name=?;`, i, strings.ToUpper(n))
}

func (d *Db) RemoveTask(n string) (result sql.Result, err error) {
	if n == "" {
		return nil, errEmptySymbol
	}

	return d.Exec(`delete from session where task_name=?;`, strings.ToUpper(n))
}

func (d *Db) GetSession() (map[string]int64, error) {
	rows, errQuery := d.Query(`select task_name,"interval" from session`)
	if errQuery != nil {
		return nil, errQuery
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
		if err := rows.Scan(&n, &i); err != nil {
			return nil, err
		}
		tasks[n] = i
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}
