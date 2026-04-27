package dbutil

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// DB はリポジトリが使うメソッドだけを定義したインターフェース
// *sql.DB も *LoggingDB も両方これを満たす
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// LoggingDB は sql.DB をラップしてクエリログを出力する
type LoggingDB struct {
	*sql.DB
}

// Wrap は sql.DB をログ付きラッパーで包む
func Wrap(db *sql.DB) *LoggingDB {
	return &LoggingDB{DB: db}
}

// ExecContext は INSERT/UPDATE/DELETE のログを出力
func (l *LoggingDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := l.DB.ExecContext(ctx, query, args...)
	logQuery("exec", query, time.Since(start), err)
	return result, err
}

// QueryRowContext は SELECT（1行）のログを出力
func (l *LoggingDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := l.DB.QueryRowContext(ctx, query, args...)
	logQuery("query", query, time.Since(start), nil)
	return row
}

// QueryContext は SELECT（複数行）のログを出力
func (l *LoggingDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := l.DB.QueryContext(ctx, query, args...)
	logQuery("query", query, time.Since(start), err)
	return rows, err
}

func logQuery(op, query string, duration time.Duration, err error) {
	if err != nil {
		slog.Error("query failed", "op", op, "sql", truncate(query, 200), "duration", duration, "error", err)
	} else {
		slog.Info("query", "op", op, "sql", truncate(query, 200), "duration", duration)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
