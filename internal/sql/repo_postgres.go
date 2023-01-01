package sql

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	psql "github.com/hyphengolang/prelude/sql/postgres"
)

type PSQLHandler[T any] interface {
	Log(...any)
	Logf(string, ...any)

	Exec(ctx context.Context, stmt string, args ...any) error
	Query(ctx context.Context, stmt string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error)
	QueryRow(ctx context.Context, stmt string, scanner func(r pgx.Row) error, args ...any) error
}

type repoHandler[T any] struct {
	conn *pgxpool.Pool
}

// Exec implements SQLHandler
func (r *repoHandler[T]) Exec(ctx context.Context, stmt string, args ...any) error {
	return psql.ExecContext(ctx, r.conn, stmt, args...)
}

// Query implements SQLHandler
func (r *repoHandler[T]) Query(ctx context.Context, stmt string, scanner func(r pgx.Rows, v *T) error, args ...any) ([]T, error) {
	return psql.QueryContext(ctx, r.conn, stmt, scanner, args...)
}

// QueryRow implements SQLHandler
func (r *repoHandler[T]) QueryRow(ctx context.Context, stmt string, scanner func(r pgx.Row) error, args ...any) error {
	return psql.QueryRowContext(ctx, r.conn, stmt, scanner, args...)
}

// Log implements Service
func (*repoHandler[T]) Log(v ...any) { log.Println(v...) }

// Logf implements Service
func (*repoHandler[T]) Logf(format string, v ...any) { log.Printf(format, v...) }

func NewPSQLHandler[T any](conn *pgxpool.Pool) PSQLHandler[T] {
	return &repoHandler[T]{conn}
}
