package postgres

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

// Driver is postgres database driver implementation
type Driver struct {
	db  *sql.DB
	log *zap.Logger
}

// NewDriver is Driver constructor
func NewDriver(ctx context.Context, db *sql.DB, log *zap.Logger) *Driver {
	return &Driver{
		db:  db,
		log: log,
	}
}

func (d *Driver) Close() error {
	return d.db.Close()
}
