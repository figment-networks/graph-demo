package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/figment-networks/graph-demo/manager/store/params"

	"github.com/figment-networks/indexing-engine/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks("network", "chain_id", "version", "epoch", "height", "hash",  "time", "numtxs" ) VALUES
	($1, $2, $3, $4, $5, $6, $7, $8 )
	ON CONFLICT (network, chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	time = EXCLUDED.time,
	numtxs = EXCLUDED.numtxs`
)

// StoreBlock appends data to buffer
func (d *Driver) StoreBlock(ctx context.Context, b structs.Block) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(insertBlock, b.Network, b.ChainID, b.Version, b.Block.Epoch, b.Block.Height, b.Block.Hash, b.Block.Time, b.Block.NumberOfTransactions)
	if err != nil {
		log.Println("[DB] Rollback flushB error: ", err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

const GetBlockForTimeQuery = `SELECT id, epoch, height, hash, time, numtxs
							FROM public.blocks
							WHERE
								chain_id = $1 AND
								network = $2 AND
								time %s $3
							ORDER BY time ASC
							LIMIT 1`

// GetBlockForTime returns first block that comes on or after given time. If no such block exists, returns closest block that comes before given time.
func (d *Driver) GetBlockForTime(ctx context.Context, blx structs.Block, time time.Time) (out structs.Block, isBefore bool, err error) {
	returnBlx := structs.Block{}

	row := d.db.QueryRowContext(ctx, fmt.Sprintf(GetBlockForTimeQuery, ">="), blx.ChainID, blx.Network, time)
	if row == nil {
		return out, isBefore, params.ErrNotFound
	}

	err = row.Scan(&returnBlx.ID, &returnBlx.Epoch, &returnBlx.Height, &returnBlx.Hash, &returnBlx.Time, &returnBlx.NumberOfTransactions)
	if err != sql.ErrNoRows {
		return returnBlx, isBefore, err
	}

	isBefore = true
	row = d.db.QueryRowContext(ctx, fmt.Sprintf(GetBlockForTimeQuery, "<"), blx.ChainID, blx.Network, time)
	if row == nil {
		return out, isBefore, params.ErrNotFound
	}
	err = row.Scan(&returnBlx.ID, &returnBlx.Epoch, &returnBlx.Height, &returnBlx.Hash, &returnBlx.Time, &returnBlx.NumberOfTransactions)
	if err == sql.ErrNoRows {
		return returnBlx, isBefore, params.ErrNotFound
	}

	return returnBlx, isBefore, err
}
