package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks("chain_id", "height", "hash", "time", "header", "data", "evidence", "last_commit", "numtxs") VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash,
	time = EXCLUDED.time,
	header = EXCLUDED.header,
	data = EXCLUDED.data,
	evidence = EXCLUDED.evidence,
	last_commit = EXCLUDED.last_commit,
	numtxs = EXCLUDED.numtxs
	`
)

// StoreBlock appends data to buffer
func (d *Driver) StoreBlock(ctx context.Context, b structs.Block) error {

	header, err := getJsonValue(b.Header)
	if err != nil {
		return err
	}

	data, err := getJsonValue(b.Data)
	if err != nil {
		return err
	}

	evidence, err := getJsonValue(b.Evidence)
	if err != nil {
		return err
	}

	lastCommit, err := getJsonValue(&b.LastCommit)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(insertBlock, b.ChainID, b.Height, b.Hash, b.Time, header, data, evidence, lastCommit, b.NumberOfTransactions)
	return err
}

func getJsonValue(v interface{}) (string, error) {
	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)

	if v == nil {
		return "", nil
	}

	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return (fmt.Stringer)(buff).String(), nil
}

func (d *Driver) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (b structs.Block, err error) {
	row := d.db.QueryRowContext(ctx, `SELECT chain_id, height, hash, time, header, data, evidence, last_commit, numtxs FROM public.blocks WHERE chain_id = $1 AND height = $2`, chainID, height)
	if row == nil {
		return b, sql.ErrNoRows
	}

	err = row.Scan(&b.ChainID, &b.Height, &b.Hash, &b.Time, &b.Header, &b.Data, &b.Evidence, &b.LastCommit, &b.NumberOfTransactions)
	if err != sql.ErrNoRows {
		return b, err
	}

	return b, nil
}

func (d *Driver) GetLatestHeight(ctx context.Context, chainID string) (height uint64, err error) {
	row := d.db.QueryRowContext(ctx, `SELECT height FROM public.progress WHERE chain_id = $1 `, chainID)
	if row == nil {
		return 0, nil
	}

	err = row.Scan(&height)
	if err != sql.ErrNoRows {
		return height, err
	}

	return height, nil
}

func (d *Driver) SetLatestHeight(ctx context.Context, chainID string, height uint64) (err error) {
	_, err = d.db.ExecContext(ctx, `INSERT INTO public.progress("chain_id", "height") VALUES ($1, $2) ON CONFLICT (chain_id) DO UPDATE SET height = EXCLUDED.height`, chainID, height)
	return err
}
