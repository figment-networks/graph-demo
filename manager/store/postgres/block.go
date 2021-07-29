package postgres

import (
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

	header, err := json.Marshal(b.Header)
	if err != nil {
		return err
	}

	data, err := json.Marshal(b.Data)
	if err != nil {
		return err
	}

	evidence, err := json.Marshal(b.Evidence)
	if err != nil {
		return err
	}

	lastCommit, err := json.Marshal(*b.LastCommit)
	if err != nil {
		return err
	}

	_, err = d.db.Exec(insertBlock, b.Header.ChainID, b.Header.Height, b.Hash, b.Header.Time, header, data, evidence, lastCommit, len(b.Data.Txs))
	return err
}

func (d *Driver) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (b structs.Block, err error) {

	row := d.db.QueryRowContext(ctx, `SELECT hash, header, data, evidence, last_commit FROM public.blocks WHERE chain_id = $1 AND height = $2`, chainID, height)

	var (
		header []byte
		data   []byte
		ev     []byte
		lc     []byte
	)

	if err = row.Scan(&b.Hash, &header, &data, &ev, &lc); err != nil {
		return b, fmt.Errorf("%s, height: %d", err.Error(), height)
	}

	if err = json.Unmarshal(header, &b.Header); err != nil {
		return b, err
	}

	if err = json.Unmarshal(data, &b.Data); err != nil {
		return b, err
	}

	if err = json.Unmarshal(ev, &b.Evidence); err != nil {
		return b, err
	}

	if lc != nil {
		b.LastCommit = &structs.Commit{}
		if err = json.Unmarshal(lc, b.LastCommit); err != nil {
			return b, err
		}
	}

	return b, err
}

func (d *Driver) GetLatestHeight(ctx context.Context, chainID string) (height uint64, err error) {
	row := d.db.QueryRowContext(ctx, `SELECT height FROM public.progress WHERE chain_id = $1`, chainID)
	if err = row.Scan(&height); err != sql.ErrNoRows {
		return height, err
	}

	return height, nil
}

func (d *Driver) SetLatestHeight(ctx context.Context, chainID string, height uint64) (err error) {
	_, err = d.db.ExecContext(ctx, `INSERT INTO public.progress("chain_id", "height") VALUES ($1, $2) ON CONFLICT (chain_id) DO UPDATE SET height = EXCLUDED.height`, chainID, height)
	return err
}
