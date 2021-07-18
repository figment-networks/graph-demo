package postgres

import (
	"context"
	"database/sql"

	"github.com/figment-networks/graph-demo/manager/structs"
)

// StoreTransactions adds transactions to storage buffer
func (d *Driver) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {

	for _, tx := range txs {

		if err := storeTx(ctx, d.db, tx); err != nil {
			return err
		}

	}
	return nil
}

func storeTx(ctx context.Context, db *sql.DB, t structs.Transaction) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// TODO(lukanus): store  REAL transation
	/*
		_, err = tx.Exec(txInsert, af.Network, t.ChainID, t.Epoch, t.Height, t.Hash, t.BlockHash, t.Time,
			pq.Array(af.Types), pq.Array(af.Parties), pq.Array(af.Senders), pq.Array(af.Recipients), pq.Array([]float64{.0}),
			fee, t.GasWanted, t.GasUsed, strings.Map(removeCharacters, t.Memo), events.String(), t.Raw, t.RawLog,
			strings.Map(removeCharacters, t.Memo)+" "+strings.Join(af.Parties, " "))
	*/
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

// GetTransactions gets transactions based on given criteria the order is forced to be time DESC
func (d *Driver) GetTransactions(ctx context.Context) (txs []structs.Transaction, err error) {
	// TODO(lukanus): Create real transactions
	return txs, nil
}
