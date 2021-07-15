package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/figment-networks/graph-demo/manager/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks("network", "chain_id", "epoch", "height", "hash",  "time", "numtxs" ) VALUES
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
	_, err = tx.Exec(insertBlock, b.Network, b.ChainID, b.Epoch, b.Height, b.Hash, b.Time, b.NumberOfTransactions)
	if err != nil {
		log.Println("[DB] Rollback flushB error: ", err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

const GetBlockByHeight = `SELECT id, epoch, height, hash, time, numtxs
							FROM public.blocks
							WHERE
								chain_id = $1 AND
								network = $2 AND
								height = $3`

// GetBlockForTime returns first block that comes on or after given time. If no such block exists, returns closest block that comes before given time.
func (d *Driver) GetBlockBytHeight(ctx context.Context, height uint64, chainID, network string) (block structs.Block, err error) {
	row := d.db.QueryRowContext(ctx, GetBlockByHeight, chainID, network, height)
	if row == nil {
		return structs.Block{}, sql.ErrNoRows
	}

	err = row.Scan(&block.ID, &block.Epoch, &block.Height, &block.Hash, &block.Time, &block.NumberOfTransactions)
	if err != sql.ErrNoRows {
		return structs.Block{}, err
	}

	return block, nil
}
