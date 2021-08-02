package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/figment-networks/graph-demo/manager/structs"
	"github.com/lib/pq"
)

const (
	txInsert = `INSERT INTO public.transactions("chain_id", "height", "hash", "block_hash", "time", "code_space", "code", 
	"result", "logs", "info", "tx_raw", "messages", "extension_options", "non_critical_extension_options", "auth_info", 
	"signatures", "gas_wanted", "gas_used", "memo", "raw_log") VALUES
	($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
	ON CONFLICT (chain_id, hash, height)
	DO UPDATE SET
	height = EXCLUDED.height,
	hash = EXCLUDED.hash,
	block_hash = EXCLUDED.block_hash,
	time = EXCLUDED.time,
	code_space = EXCLUDED.code_space,
	code = EXCLUDED.code,
	result = EXCLUDED.result,
	logs = EXCLUDED.logs,
	info = EXCLUDED.info,
	tx_raw = EXCLUDED.tx_raw,
	messages = EXCLUDED.messages,
	extension_options = EXCLUDED.extension_options,
	non_critical_extension_options = EXCLUDED.non_critical_extension_options,
	auth_info = EXCLUDED.auth_info,
	signatures = EXCLUDED.signatures,
	gas_wanted = EXCLUDED.gas_wanted,
	gas_used = EXCLUDED.gas_used,
	memo = EXCLUDED.memo,
	raw_log = EXCLUDED.raw_log`

	txSelect = `SELECT hash, block_hash, time, code_space, code, result, logs, info, tx_raw, messages, extension_options, 
	non_critical_extension_options, auth_info, signatures, gas_wanted, gas_used, memo, raw_log
	FROM public.transactions 
	WHERE chain_id = $1 AND height = $2`
)

// StoreTransactions adds transactions to storage buffer
func (d *Driver) StoreTransactions(ctx context.Context, txs []structs.Transaction) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}

	for _, t := range txs {

		mf, err := getMarshaledFields(t)
		if err != nil {
			return err
		}

		if err := d.storeTx(ctx, t, mf); err != nil {
			return err
		}

	}

	return tx.Commit()
}

type marshaledFields struct {
	authInfo                    []byte
	extensionOptions            []byte
	logs                        []byte
	messages                    []byte
	nonCriticalExtensionOptions []byte
	txRaw                       []byte
}

func getMarshaledFields(tx structs.Transaction) (mf marshaledFields, err error) {
	if tx.AuthInfo != nil {
		if mf.authInfo, err = json.Marshal(&tx.AuthInfo); err != nil {
			return mf, err
		}
	}

	if mf.logs, err = json.Marshal(tx.Logs); err != nil {
		return mf, err
	}

	if mf.txRaw, err = json.Marshal(tx.TxRaw); err != nil {
		return mf, err
	}

	if mf.extensionOptions, err = json.Marshal(tx.ExtensionOptions); err != nil {
		return mf, err
	}

	if mf.nonCriticalExtensionOptions, err = json.Marshal(tx.NonCriticalExtensionOptions); err != nil {
		return mf, err
	}

	if mf.messages, err = json.Marshal(tx.Messages); err != nil {
		return mf, err
	}

	// mf.extensionOptions = make([][]byte, len(tx.ExtensionOptions))
	// for i, extensionOption := range tx.ExtensionOptions {
	// 	if mf.extensionOptions[i], err = json.Marshal(extensionOption); err != nil {
	// 		return mf, err
	// 	}
	// }

	// mf.nonCriticalExtensionOptions = make([][]byte, len(tx.NonCriticalExtensionOptions))
	// for i, nonCriticalExtensionOption := range tx.NonCriticalExtensionOptions {
	// 	if mf.nonCriticalExtensionOptions[i], err = json.Marshal(nonCriticalExtensionOption); err != nil {
	// 		return mf, err
	// 	}
	// }

	// mf.messages = make([][]byte, len(tx.Messages))
	// for i, message := range tx.Messages {
	// 	if mf.messages[i], err = json.Marshal(message); err != nil {
	// 		return mf, err
	// 	}
	// }

	return
}

func (d *Driver) storeTx(ctx context.Context, t structs.Transaction, mf marshaledFields) (err error) {

	// TODO(lukanus): store  REAL transation
	_, err = d.db.ExecContext(ctx, txInsert, t.ChainID, t.Height, t.Hash, t.BlockHash, t.Time, t.CodeSpace, t.Code, t.Result, mf.logs,
		t.Info, mf.txRaw, mf.messages, mf.extensionOptions, mf.nonCriticalExtensionOptions, mf.authInfo, pq.Array(t.Signatures),
		t.GasWanted, t.GasUsed, t.Memo, t.RawLog)

	return
}

// Removes ASCII hex 0-7 causing utf-8 error in db
func removeCharacters(r rune) rune {
	if r < 7 {
		return -1
	}
	return r
}

// GetTransactions gets transactions based on given criteria the order is forced to be time DESC
func (d *Driver) GetTransactionsByHeight(ctx context.Context, height uint64, chainID string) (txs []structs.Transaction, err error) {
	rows, err := d.db.QueryContext(ctx, txSelect, chainID, height)
	if err != nil {
		return nil, err
	}

	if rows == nil {
		return nil, sql.ErrNoRows
	}

	// br := &bytes.Reader{}
	// feeDec := json.NewDecoder(br)

	for rows.Next() {
		var authInfoBytes, eoBytes, logsBytes, msgsBytes, nceoBytes, txRawBytes []byte
		var signatures pq.StringArray
		tx := structs.Transaction{
			Height:  height,
			ChainID: chainID,
		}

		err = rows.Scan(&tx.Hash, &tx.BlockHash, &tx.Time, &tx.CodeSpace, &tx.Code, &tx.Result, &logsBytes,
			&tx.Info, &txRawBytes, &msgsBytes, &eoBytes, &nceoBytes, &authInfoBytes, &signatures,
			&tx.GasWanted, &tx.GasUsed, &tx.Memo, &tx.RawLog)
		if err != nil {
			return nil, err
		}

		tx.Signatures = make([]string, len(signatures))
		for i, signature := range signatures {
			tx.Signatures[i] = signature
		}

		if err = json.Unmarshal(authInfoBytes, &tx.AuthInfo); err != nil {
			return txs, err
		}

		if err = json.Unmarshal(logsBytes, &tx.Logs); err != nil {
			return txs, err
		}

		if err = json.Unmarshal(msgsBytes, &tx.Messages); err != nil {
			return txs, err
		}

		if err = json.Unmarshal(eoBytes, &tx.ExtensionOptions); err != nil {
			return txs, err
		}

		if err = json.Unmarshal(nceoBytes, &tx.NonCriticalExtensionOptions); err != nil {
			return txs, err
		}

		if err = json.Unmarshal(txRawBytes, &tx.TxRaw); err != nil {
			return txs, err
		}

		txs = append(txs, tx)
	}

	return txs, nil
}
