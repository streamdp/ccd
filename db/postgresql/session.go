package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var (
	errEmptyTaskName = errors.New("empty task name")
	errEmptySymbol   = errors.New("empty symbol")
)

func (d *Db) AddTask(ctx context.Context, n string, i int64) (sql.Result, error) {
	if n == "" {
		return nil, errEmptyTaskName
	}

	result, err := d.ExecContext(ctx,
		`insert into session (task_name,interval) values ($1,$2) on conflict do nothing;`, strings.ToUpper(n), i,
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

	result, err := d.ExecContext(ctx, `update session set interval=$2 where task_name=$1;`, strings.ToUpper(n), i)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) RemoveTask(ctx context.Context, n string) (sql.Result, error) {
	if n == "" {
		return nil, errEmptySymbol
	}

	result, err := d.ExecContext(ctx, `delete from session where task_name=$1;`, strings.ToUpper(n))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
	}

	return result, nil
}

func (d *Db) GetSession(ctx context.Context) (map[string]int64, error) {
	rows, err := d.QueryContext(ctx, `select task_name,"interval" from session`)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errExecuteQuery, err)
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
			return nil, fmt.Errorf("%w: %w", errCopyResult, err)
		}
		tasks[n] = i
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%w: %w", errParseResults, rows.Err())
	}

	return tasks, nil
}
