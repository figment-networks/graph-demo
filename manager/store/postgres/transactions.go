package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	txInsert = `INSERT INTO public.transactions("chain_id", "height", "hash", "block_hash", "time", "fee", "gas_wanted", "gas_used", "memo", "data", "raw", "raw_log", "has_error", "type", "parties", "senders", "recipients") VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	ON CONFLICT ( chain_id, hash)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash,
	time = EXCLUDED.time,
	block_hash = EXCLUDED.block_hash,
	fee = EXCLUDED.fee,
	gas_wanted = EXCLUDED.gas_wanted,
	gas_used = EXCLUDED.gas_used,
	memo = EXCLUDED.memo,
	data = EXCLUDED.data,
	raw = EXCLUDED.raw,
	raw_log = EXCLUDED.raw_log,
	has_error = EXCLUDED.has_error,
	type = EXCLUDED.type,
	parties = EXCLUDED.parties,
	senders = EXCLUDED.senders,
	recipients = EXCLUDED.recipients`
)

// StoreTransactions adds transactions to storage buffer
func (d *Driver) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	var fee []byte
	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	for _, t := range txs {

		if err := enc.Encode(t.Events); err != nil {
			return err
		}

		af := getTransactionAdditionalFields(t)

		if t.Fee != nil {
			fee, err = json.Marshal(t.Fee)
			if err != nil {
				fee = []byte("{}")
			}
		} else {
			fee = []byte("{}")
		}

		if err := d.storeTx(ctx, d.db, t, fee, buff, af); err != nil {
			return err
		}

	}

	return tx.Commit()
}

type additionalFields struct {
	types      []string
	parties    []string
	recipients []string
	senders    []string
}

func getTransactionAdditionalFields(tx structs.Transaction) (af additionalFields) {
	for _, ev := range tx.Events {
		for _, sub := range ev.Sub {
			if len(sub.Recipient) > 0 {
				af.parties = uniqueEntriesEvTransfer(sub.Recipient, af.parties)
				af.recipients = uniqueEntriesEvTransfer(sub.Recipient, af.recipients)
			}
			if len(sub.Sender) > 0 {
				af.parties = uniqueEntriesEvTransfer(sub.Sender, af.parties)
				af.senders = uniqueEntriesEvTransfer(sub.Sender, af.senders)
			}

			if len(sub.Node) > 0 {
				for _, accounts := range sub.Node {
					af.parties = uniqueEntriesAccount(accounts, af.parties)
				}
			}

			if sub.Error != nil {
				af.types = uniqueEntry("error", af.types)
			}

			af.types = uniqueEntries(sub.Type, af.types)
		}
		af.types = uniqueEntries(ev.Type, af.types)
	}

	return
}

func uniqueEntriesEvTransfer(in []structs.EventTransfer, out []string) []string {
	for _, r := range in { // (lukanus): faster than a map :)
		var exists bool
	Inner:
		for _, re := range out {
			if r.Account.ID == re {
				exists = true
				break Inner
			}
		}
		if !exists {
			out = append(out, r.Account.ID)
		}
	}
	return out
}

func uniqueEntriesAccount(in []structs.Account, out []string) []string {
	for _, r := range in { // (lukanus): faster than a map :)
		var exists bool
	Inner:
		for _, re := range out {
			if r.ID == re {
				exists = true
				break Inner
			}
		}
		if !exists {
			out = append(out, r.ID)
		}
	}
	return out
}

func uniqueEntry(in string, out []string) []string {
	if in == "" {
		return out
	}
	for _, re := range out {
		if in == re {
			return out
		}
	}
	return append(out, in)
}

func uniqueEntries(in, out []string) []string {
	for _, r := range in { // (lukanus): faster than a map :)
		if r == "" {
			continue
		}
		var exists bool
	Inner:
		for _, re := range out {
			if r == re {
				exists = true
				break Inner
			}
		}
		if !exists {
			out = append(out, r)
		}
	}
	return out
}

func (d *Driver) storeTx(ctx context.Context, db *sql.DB, t structs.Transaction, fee []byte, events fmt.Stringer, af additionalFields) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// TODO(lukanus): store  REAL transation
	_, err = tx.Exec(txInsert, t.ChainID, t.Height, t.Hash, t.BlockHash, t.Time, fee, t.GasWanted, t.GasUsed, t.Memo, events.String(),
		t.Raw, t.RawLog, t.HasErrors, pq.Array(af.types), pq.Array(af.parties), pq.Array(af.senders), pq.Array(af.recipients))
	if err != nil {
		d.log.Error("[DB] Rollback flushB error: ", zap.Error(err))
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Removes ASCII hex 0-7 causing utf-8 error in db
func removeCharacters(r rune) rune {
	if r < 7 {
		return -1
	}
	return r
}

const GetTransactionsByHeight = `SELECT chain_id, height, hash, block_hash, time, fee, gas_wanted, gas_used, memo, events, raw, has_error
							FROM public.transactions
							WHERE chain_id = $1 AND height = $2`

// GetTransactions gets transactions based on given criteria the order is forced to be time DESC
func (d *Driver) GetTransactions(ctx context.Context, height uint64, chainID string) (txs []structs.Transaction, err error) {
	rows, err := d.db.QueryContext(ctx, GetTransactionsByHeight, chainID, height)
	if err != nil {
		return nil, err
	}

	if rows == nil {
		return nil, sql.ErrNoRows
	}

	br := &bytes.Reader{}
	feeDec := json.NewDecoder(br)

	for rows.Next() {
		var tx structs.Transaction

		byteFee := []byte{}

		err = rows.Scan(&tx.ChainID, &tx.Height, &tx.Hash, &tx.BlockHash, &tx.Time, &byteFee, &tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.Events, &tx.Raw, &tx.HasErrors)
		if err != sql.ErrNoRows {
			return nil, err
		}

		br.Reset(byteFee)
		feeDec.Decode(&tx.Fee)
		byteFee = nil

		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
	}

	return txs, nil
}
