package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{
		pool: pool,
	}
}

func (m *TxManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	if m.txFromContext(ctx) != nil {
		return fn(ctx)
	}

	err := pgx.BeginTxFunc(ctx, m.pool, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	}, func(tx pgx.Tx) error {
		return fn(m.contextWithTx(ctx, tx))
	})
	if err != nil {
		return fmt.Errorf("execute read-only PostgreSQL transaction: %w", err)
	}
	return nil
}

func (m *TxManager) Executor(ctx context.Context) Executor {
	if tx := m.txFromContext(ctx); tx != nil {
		return tx
	}
	return m.pool
}
