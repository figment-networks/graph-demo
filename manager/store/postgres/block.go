package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/figment-networks/graph-demo/manager/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks( "chain_id", "height", "hash", "time") VALUES
	($1, $2, $3, $4)
	ON CONFLICT ( chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash
	time = EXCLUDED.time
	`
)

// StoreBlock appends data to buffer
func (d *Driver) StoreBlock(ctx context.Context, b structs.Block) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(insertBlock, b.ChainID, b.Height, b.Hash, b.Time)
	if err != nil {
		log.Println("[DB] Rollback flushB error: ", err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

const GetBlockByHeight = `SELECT hash, height, time, chain_id
							FROM public.blocks
							WHERE chain_id = $1 AND height = $2`

// GetBlockForTime returns first block that comes on or after given time. If no such block exists, returns closest block that comes before given time.
func (d *Driver) GetBlockBytHeight(ctx context.Context, height uint64, chainID string) (b structs.Block, err error) {
	row := d.db.QueryRowContext(ctx, GetBlockByHeight, chainID, height)
	if row == nil {
		return structs.Block{}, sql.ErrNoRows
	}

	err = row.Scan(&b.Hash, &b.Height, &b.Time, &b.ChainID)
	if err != sql.ErrNoRows {
		return structs.Block{}, err
	}

	return b, nil
}
