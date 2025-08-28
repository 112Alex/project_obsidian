package database

import (
	"context"
	"fmt"

	"github.com/112Alex/project_obsidian/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB представляет собой обертку над пулом соединений PostgreSQL
type PostgresDB struct {
	pool *pgxpool.Pool
}

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(ctx context.Context, cfg config.PostgresConfig) (*PostgresDB, error) {
	// Создание конфигурации пула соединений
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	// Установка максимального размера пула соединений
	poolConfig.MaxConns = int32(cfg.PoolMax)

	// Создание пула соединений
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

// Close закрывает соединение с базой данных
func (db *PostgresDB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Pool возвращает пул соединений
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Exec выполняет SQL-запрос без возврата результатов
func (db *PostgresDB) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, sql, args...)
}

// Query выполняет SQL-запрос и возвращает результаты
func (db *PostgresDB) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, sql, args...)
}

// QueryRow выполняет SQL-запрос и возвращает одну строку результата
func (db *PostgresDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, sql, args...)
}

// Begin начинает новую транзакцию
func (db *PostgresDB) Begin(ctx context.Context) (pgx.Tx, error) {
	return db.pool.Begin(ctx)
}
