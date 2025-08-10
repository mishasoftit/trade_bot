package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TimescaleDB represents a TimescaleDB connection
type TimescaleDB struct {
	pool *pgxpool.Pool
}

// ConnectTimescaleDB establishes a connection to TimescaleDB
func ConnectTimescaleDB(ctx context.Context) (*TimescaleDB, error) {
	// TODO: Load from config
	connStr := "postgres://user:password@localhost:5432/trading?sslmode=disable"

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return &TimescaleDB{pool: pool}, nil
}

// Close closes the database connection
func (db *TimescaleDB) Close(ctx context.Context) {
	db.pool.Close()
}

// LogOrder logs an order to the database
func (db *TimescaleDB) LogOrder(o *order.Order) error {
	query := `
		INSERT INTO orders (
			id, symbol, type, side, price, quantity, exchange, status, 
			created_at, updated_at, retry_count
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := db.pool.Exec(context.Background(), query,
		o.ID, o.Symbol, int(o.Type), int(o.Side), o.Price, o.Quantity, 
		o.Exchange, int(o.Status), o.CreatedAt, o.UpdatedAt, o.RetryCount,
	)

	return err
}

// LogTrade logs a trade execution to the database
func (db *TimescaleDB) LogTrade(t *Trade) error {
	query := `
		INSERT INTO trades (
			order_id, exchange_id, symbol, price, quantity, fee, fee_currency,
			executed_at, side
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	_, err := db.pool.Exec(context.Background(), query,
		t.OrderID, t.ExchangeID, t.Symbol, t.Price, t.Quantity, 
		t.Fee, t.FeeCurrency, t.ExecutedAt, int(t.Side),
	)

	return err
}

// CreateSchema creates the necessary tables if they don't exist
func (db *TimescaleDB) CreateSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			symbol TEXT NOT NULL,
			type SMALLINT NOT NULL,
			side SMALLINT NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			exchange TEXT NOT NULL,
			status SMALLINT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			retry_count INTEGER NOT NULL
		)`,
		
		`SELECT create_hypertable('orders', 'created_at', if_not_exists => TRUE)`,
		
		`CREATE TABLE IF NOT EXISTS trades (
			id SERIAL PRIMARY KEY,
			order_id TEXT NOT NULL REFERENCES orders(id),
			exchange_id TEXT NOT NULL,
			symbol TEXT NOT NULL,
			price DOUBLE PRECISION NOT NULL,
			quantity DOUBLE PRECISION NOT NULL,
			fee DOUBLE PRECISION,
			fee_currency TEXT,
			executed_at TIMESTAMPTZ NOT NULL,
			side SMALLINT NOT NULL
		)`,
		
		`SELECT create_hypertable('trades', 'executed_at', if_not_exists => TRUE)`,
		
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)`,
	}

	for _, query := range queries {
		_, err := db.pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("error executing query %q: %w", query, err)
		}
	}

	return nil
}

// Trade represents a trade execution
type Trade struct {
	OrderID     string
	ExchangeID  string
	Symbol      string
	Price       float64
	Quantity    float64
	Fee         float64
	FeeCurrency string
	ExecutedAt  time.Time
	Side        order.Side
}