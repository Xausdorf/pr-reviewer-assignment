package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type errRow struct {
	err error
}

func (e errRow) Scan(dest ...any) error {
	return e.err
}

type errRows struct {
	err error
}

func (errRows) Close() {
}
func (e errRows) Err() error {
	return e.err
}
func (errRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}
func (errRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}
func (errRows) Next() bool {
	return false
}
func (e errRows) Scan(dest ...any) error {
	return e.err
}
func (e errRows) Values() ([]any, error) {
	return nil, e.err
}
func (e errRows) RawValues() [][]byte {
	return nil
}
func (e errRows) Conn() *pgx.Conn {
	return nil
}

type toSqler interface {
	ToSql() (string, []interface{}, error)
}

type execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func tryExec(query toSqler, executor execer, ctx context.Context) (pgconn.CommandTag, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return executor.Exec(ctx, sql, args...)
}

func tryQueryRow(query toSqler, pool *pgxpool.Pool, ctx context.Context) pgx.Row {
	sql, args, err := query.ToSql()
	if err != nil {
		return errRow{err: err}
	}
	return pool.QueryRow(ctx, sql, args...)
}

func tryQuery(query toSqler, pool *pgxpool.Pool, ctx context.Context) (pgx.Rows, error) {
	sql, args, err := query.ToSql()
	if err != nil {
		return errRows{err: err}, err
	}
	return pool.Query(ctx, sql, args...)
}
