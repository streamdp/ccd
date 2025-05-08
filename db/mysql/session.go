package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var errEmptyTaskName = errors.New("empty task name")

func (d *Db) AddTask(ctx context.Context, n string, i int64) (sql.Result, error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	result, err := d.ExecContext(
		ctx, "insert ignore into session (task_name,session.interval) values (?,?);", strings.ToUpper(n), i,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) UpdateTask(ctx context.Context, n string, i int64) (sql.Result, error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	result, err := d.ExecContext(
		ctx, "update session set session.interval=? where task_name=?;", i, strings.ToUpper(n),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) RemoveTask(ctx context.Context, n string) (sql.Result, error) {
	if n == "" {
		return nil, errEmptySymbol
	}

	result, err := d.ExecContext(ctx, `delete from session where task_name=?;`, strings.ToUpper(n))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) GetSession(ctx context.Context) (map[string]int64, error) {
	rows, errQuery := d.QueryContext(ctx, `select task_name,session.interval from session`)
	if errQuery != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, errQuery)
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
			return nil, fmt.Errorf("%w: %w", errCopyResult, err)
		}
		tasks[n] = i
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errParseResults, err)
	}

	return tasks, nil
}
