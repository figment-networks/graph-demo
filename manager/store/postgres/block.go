package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"
)

const (
	insertBlock = `INSERT INTO public.blocks("chain_id", "height", "hash", "time", "header", "data", "evidence", "last_commit") VALUES
	($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash,
	time = EXCLUDED.time,
	header = EXCLUDED.header,
	data = EXCLUDED.data,
	evidence = EXCLUDED.evidence,
	last_commit = EXCLUDED.last_commit
	RETURNING id`

	selectBlock = `SELECT hash, time, header, data, evidence, last_commit FROM public.blocks WHERE chain_id = $1 AND height = $2`
)

// StoreBlock appends data to buffer
func (d *Driver) StoreBlock(ctx context.Context, b structs.Block) (string, error) {

	header, err := json.Marshal(b.Header)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(b.Data)
	if err != nil {
		return "", err
	}

	evidence, err := json.Marshal(b.Evidence)
	if err != nil {
		return "", err
	}

	lastCommit, err := json.Marshal(*b.LastCommit)
	if err != nil {
		return "", err
	}

	var id string
	row := d.db.QueryRow(insertBlock, b.ChainID, b.Height, b.Hash, b.Time, header, data, evidence, lastCommit).Scan(&id)
	if row != nil {
		return "", errors.New(row.Error())
	}

	return id, err
}

func (d *Driver) GetBlockByHeight(ctx context.Context, height uint64, chainID string) (b structs.Block, err error) {

	row := d.db.QueryRowContext(ctx, selectBlock, chainID, height)

	var (
		header []byte
		data   []byte
		ev     []byte
		lc     []byte
	)

	b = structs.Block{
		Height:  height,
		ChainID: chainID,
	}

	if err = row.Scan(&b.Hash, &b.Time, &header, &data, &ev, &lc); err != nil {
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
