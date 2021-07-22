package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/figment-networks/graph-demo/manager/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks("chain_id", "height", "hash", "time", "header", "data", "evidence", "last_commit", "tx_num") VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT ( chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash
	time = EXCLUDED.time
	header = EXCLUDED.header
	data = EXCLUDED.data
	evidence = EXCLUDED.evidence
	last_commit = EXCLUDED.last_commit
	tx_num = EXCLUDED.tx_num
	`
)

// StoreBlock appends data to buffer
func (d *Driver) StoreBlock(ctx context.Context, b structs.Block) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)

	if err := enc.Encode(b.Header); err != nil {
		return err
	}
	header := buff.String()

	if err := enc.Encode(b.Data); err != nil {
		return err
	}
	data := buff.String()

	if err := enc.Encode(b.Evidence); err != nil {
		return err
	}
	evidence := buff.String()

	if err := enc.Encode(b.LastCommit); err != nil {
		return err
	}
	lastCommit := buff.String()

	_, err = tx.Exec(insertBlock, b.ChainID, b.Height, b.Hash, b.Time, header, data, evidence, lastCommit)
	if err != nil {
		log.Println("[DB] Rollback flushB error: ", err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

const GetBlockByHeight = `SELECT chain_id, height, hash, time, header, data, evidence, last_commit, tx_num
							FROM public.blocks
							WHERE chain_id = $1 AND height = $2`

// GetBlockForTime returns first block that comes on or after given time. If no such block exists, returns closest block that comes before given time.
func (d *Driver) GetBlockBytHeight(ctx context.Context, height uint64, chainID string) (b structs.Block, err error) {
	row := d.db.QueryRowContext(ctx, GetBlockByHeight, chainID, height)
	if row == nil {
		return structs.Block{}, sql.ErrNoRows
	}

	err = row.Scan(&b.ChainID, &b.Height, &b.Hash, &b.Time, &b.Header, &b.Data, &b.Evidence, &b.LastCommit, &b.NumberOfTransactions)
	if err != sql.ErrNoRows {
		return structs.Block{}, err
	}

	return b, nil
}
