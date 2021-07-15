package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/figment-networks/graph-demo/manager/store/params"
	"github.com/figment-networks/graph-demo/manager/structs"

	"github.com/lib/pq"
)

type accountSearch string

const asBoth accountSearch = "Both"
const asOnlyArray accountSearch = "OnlyArray"
const accountSearchMode = asBoth

type AdditionalFields struct {
	Parties    []string
	Senders    []string
	Recipients []string
	Types      []string

	Network string
}

// StoreTransactions adds transactions to storage buffer
func (d *Driver) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {

	buff := &bytes.Buffer{}
	enc := json.NewEncoder(buff)
	var err error
	for _, tx := range txs {
		af := AdditionalFields{
			Network: tx.Network,
		}

		// (lukanus): candidate for goroutine
		for _, ev := range tx.Events {
			for _, sub := range ev.Sub {
				if len(sub.Recipient) > 0 {
					af.Parties = uniqueEntriesEvTransfer(sub.Recipient, af.Parties)
					af.Recipients = uniqueEntriesEvTransfer(sub.Recipient, af.Recipients)
				}
				if len(sub.Sender) > 0 {
					af.Parties = uniqueEntriesEvTransfer(sub.Sender, af.Parties)
					af.Senders = uniqueEntriesEvTransfer(sub.Sender, af.Senders)
				}

				if len(sub.Node) > 0 {
					for _, accounts := range sub.Node {
						af.Parties = uniqueEntriesAccount(accounts, af.Parties)
					}
				}

				if sub.Error != nil {
					af.Types = uniqueEntry("error", af.Types)
				}

				af.Types = uniqueEntries(sub.Type, af.Types)
			}
			af.Types = uniqueEntries(ev.Type, af.Types)
		}

		var fee []byte
		if tx.Fee != nil {
			fee, err = json.Marshal(tx.Fee)
			if err != nil {
				fee = []byte("{}")
			}
		} else {
			fee = []byte("{}")
		}

		if err := enc.Encode(tx.Events); err != nil {
			return err
		}

		if len(af.Types) == 0 {
			log.Println("additional fields: ", af, tx.Height, tx.Hash)
		}

		if err := storeTx(ctx, d.db, tx, af, fee, buff); err != nil {
			return err
		}
		buff.Reset()
	}
	return nil
}

const (
	txInsert = `INSERT INTO public.transaction_events
			("network", "chain_id", "epoch", "height", "hash", "block_hash", "time", "type", "parties", "senders", "recipients", "amount", "fee", "gas_wanted", "gas_used", "memo", "data", "raw", "raw_log", "fulltext_col") VALUES
			( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, to_tsvector('english', $20) )
	 ON CONFLICT (network, chain_id, hash, height)
		DO UPDATE SET height = EXCLUDED.height,
		time = EXCLUDED.time,
		type = EXCLUDED.type,
		parties = EXCLUDED.parties,
		senders = EXCLUDED.senders,
		recipients = EXCLUDED.recipients,
		data = EXCLUDED.data,
		raw = EXCLUDED.raw,
		amount = EXCLUDED.amount,
		block_hash = EXCLUDED.block_hash,
		gas_wanted = EXCLUDED.gas_wanted,
		gas_used = EXCLUDED.gas_used,
		memo = EXCLUDED.memo,
		fulltext_col = to_tsvector('english', CONCAT( EXCLUDED.memo, ' ', array_to_string(EXCLUDED.parties, ' ', ' '))),
		fee = EXCLUDED.fee,
		raw_log = EXCLUDED.raw_log`
)

func storeTx(ctx context.Context, db *sql.DB, t structs.Transaction, af AdditionalFields, fee []byte, events fmt.Stringer) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(txInsert, af.Network, t.ChainID, t.Epoch, t.Height, t.Hash, t.BlockHash, t.Time,
		pq.Array(af.Types), pq.Array(af.Parties), pq.Array(af.Senders), pq.Array(af.Recipients), pq.Array([]float64{.0}),
		fee, t.GasWanted, t.GasUsed, strings.Map(removeCharacters, t.Memo), events.String(), t.Raw, t.RawLog,
		strings.Map(removeCharacters, t.Memo)+" "+strings.Join(af.Parties, " "))

	if err != nil {
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

// GetTransactions gets transactions based on given criteria the order is forced to be time DESC
func (d *Driver) GetTransactions(ctx context.Context, tsearch params.TransactionSearch) (txs []structs.Transaction, err error) {
	var i = 1

	parts := []string{}
	data := []interface{}{}

	sortType := "time"

	if tsearch.Network != "" {
		parts = append(parts, "network = $"+strconv.Itoa(i))
		data = append(data, tsearch.Network)
		i++
	}

	if len(tsearch.ChainIDs) > 0 {
		chains := "chain_id IN ("
		for j, c := range tsearch.ChainIDs {
			if j > 0 {
				chains += ","
			}
			data = append(data, c)
			chains += "$" + strconv.Itoa(i)
			i++
		}
		parts = append(parts, chains+")")
	}

	if tsearch.Epoch != "" {
		parts = append(parts, "epoch = $"+strconv.Itoa(i))
		data = append(data, tsearch.Epoch)
		i++
	}

	if tsearch.Hash != "" {
		parts = append(parts, "hash = $"+strconv.Itoa(i))
		data = append(data, tsearch.Hash)
		i++
	}

	if tsearch.BlockHash != "" {
		parts = append(parts, "block_hash = $"+strconv.Itoa(i))
		data = append(data, tsearch.BlockHash)
		i++
	}

	if tsearch.Height > 0 {
		parts = append(parts, "height = $"+strconv.Itoa(i))
		data = append(data, tsearch.Height)
		sortType = "height"
		i++
	} else {
		if tsearch.AfterHeight > 0 {
			parts = append(parts, "height > $"+strconv.Itoa(i))
			sortType = "height"
			data = append(data, tsearch.AfterHeight)
			i++
		}

		if tsearch.BeforeHeight > 0 {
			parts = append(parts, "height < $"+strconv.Itoa(i))
			sortType = "height"
			data = append(data, tsearch.BeforeHeight)
			i++
		}
	}

	// SELECT * FROM transaction_events WHERE type @> ARRAY['delegate']::varchar[] limit 1;
	if len(tsearch.Type.Value) > 0 {
		var q string = "type @> $"
		if tsearch.Type.Any {
			q = "type <@ $"
		}
		parts = append(parts, q+strconv.Itoa(i))
		data = append(data, pq.Array(tsearch.Type.Value))
		i++
	}

	if len(tsearch.Account) > 0 || len(tsearch.Sender) > 0 || len(tsearch.Receiver) > 0 {
		accountsToFetch := append(tsearch.Account, tsearch.Sender...)
		accountsToFetch = append(accountsToFetch, tsearch.Receiver...)
		if accountSearchMode != asOnlyArray {
			parts = append(parts, "fulltext_col @@ to_tsquery('english', $"+strconv.Itoa(i)+")")
			data = append(data, strings.Join(accountsToFetch, " | "))
			i++
		}
		parts = append(parts, "parties @> $"+strconv.Itoa(i))
		data = append(data, pq.Array(accountsToFetch))
		i++
	}

	if len(tsearch.Sender) > 0 {
		parts = append(parts, "senders @> $"+strconv.Itoa(i))
		data = append(data, pq.Array(tsearch.Sender))
		i++
	}

	if len(tsearch.Receiver) > 0 {
		parts = append(parts, "recipients @> $"+strconv.Itoa(i))
		data = append(data, pq.Array(tsearch.Receiver))
		i++
	}

	if len(tsearch.Memo) > 0 {
		parts = append(parts, "fulltext_col @@ to_tsquery('english', $"+strconv.Itoa(i)+")")
		data = append(data, tsearch.Memo)
		i++
	}

	if !tsearch.AfterTime.IsZero() {
		parts = append(parts, "time >= $"+strconv.Itoa(i))
		data = append(data, tsearch.AfterTime)
		sortType = "time"
		i++
	}

	if !tsearch.BeforeTime.IsZero() {
		parts = append(parts, "time <= $"+strconv.Itoa(i))
		data = append(data, tsearch.BeforeTime)
		sortType = "time"
	}

	qBuilder := strings.Builder{}
	qBuilder.WriteString("SELECT id, chain_id, epoch, fee, height, hash, block_hash, time, gas_wanted, gas_used, memo, data, (type @> '{error}') as has_errors")

	if tsearch.WithRaw {
		qBuilder.WriteString(", raw ")
	}

	if tsearch.WithRawLog {
		qBuilder.WriteString(", raw_log ")
	}

	qBuilder.WriteString(" FROM public.transaction_events WHERE ")
	for i, par := range parts {
		if i != 0 {
			qBuilder.WriteString(" AND ")
		}
		qBuilder.WriteString(par)
	}

	if sortType == "time" {
		qBuilder.WriteString(" ORDER BY time DESC")
	} else {
		qBuilder.WriteString(" ORDER BY height DESC")
	}

	if tsearch.Limit > 0 {
		qBuilder.WriteString(" LIMIT " + strconv.FormatUint(uint64(tsearch.Limit), 10))

		if tsearch.Offset > 0 {
			qBuilder.WriteString(" OFFSET " + strconv.FormatUint(uint64(tsearch.Offset), 10))
		}
	}

	a := qBuilder.String()

	rows, err := d.db.QueryContext(ctx, a, data...)
	switch {
	case err == sql.ErrNoRows:
		return nil, params.ErrNotFound
	case err != nil:
		return nil, fmt.Errorf("query error: %w", err)
	default:
	}

	defer rows.Close()

	br := &bytes.Reader{}
	feeDec := json.NewDecoder(br)

	for rows.Next() {
		tx := structs.Transaction{}
		byteFee := []byte{}

		if tsearch.WithRaw && tsearch.WithRawLog {
			err = rows.Scan(&tx.ID, &tx.ChainID, &tx.Epoch, &byteFee, &tx.Height, &tx.Hash, &tx.BlockHash, &tx.Time, &tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.Events, &tx.HasErrors, &tx.Raw, &tx.RawLog)
		} else if tsearch.WithRaw {
			err = rows.Scan(&tx.ID, &tx.ChainID, &tx.Epoch, &byteFee, &tx.Height, &tx.Hash, &tx.BlockHash, &tx.Time, &tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.Events, &tx.HasErrors, &tx.Raw)
		} else if tsearch.WithRawLog {
			err = rows.Scan(&tx.ID, &tx.ChainID, &tx.Epoch, &byteFee, &tx.Height, &tx.Hash, &tx.BlockHash, &tx.Time, &tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.Events, &tx.HasErrors, &tx.RawLog)
		} else {
			err = rows.Scan(&tx.ID, &tx.ChainID, &tx.Epoch, &byteFee, &tx.Height, &tx.Hash, &tx.BlockHash, &tx.Time, &tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.Events, &tx.HasErrors)
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
