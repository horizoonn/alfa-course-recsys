package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Executor interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ExecutorFunc func(context.Context) Executor
