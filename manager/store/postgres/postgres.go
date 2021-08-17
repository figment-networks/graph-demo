package postgres

import (
	"context"
	"database/sql"
	"time"

	"go.uber.org/zap"
)

// Driver is postgres database driver implementation
type Driver struct {
	db  *sql.DB
	log *zap.Logger
}

// NewDriver is Driver constructor
func NewDriver(ctx context.Context, log *zap.Logger, dbURL string) (*Driver, error) {
	db, err := connectPostgres(ctx, log, dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(60)
	db.SetConnMaxLifetime(time.Minute * 10)

	return &Driver{
		db:  db,
		log: log,
	}, nil
}

func connectPostgres(ctx context.Context, log *zap.Logger, dbURL string) (*sql.DB, error) {
	defer log.Sync()
	log.Info("[DB] Connecting to database...")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	log.Info("[DB] Ping successfull...")
	return db, nil
}

func (d *Driver) Close() error {
	return d.db.Close()
}
