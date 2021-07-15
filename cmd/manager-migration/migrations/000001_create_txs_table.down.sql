DROP INDEX IF EXISTS idx_tx_ev_height;
DROP INDEX IF EXISTS idx_tx_ev_time;
DROP INDEX IF EXISTS idx_tx_ev_hash;
DROP INDEX IF EXISTS idx_tx_ev_block_hash;
DROP INDEX IF EXISTS idx_tx_ev_parties_gin;
DROP INDEX IF EXISTS transaction_txts_idx;
DROP INDEX IF EXISTS transaction_types_idx;
DROP INDEX IF EXISTS transaction_senders_idx;
DROP INDEX IF EXISTS transaction_recipients_idx;

DROP TABLE IF EXISTS transaction_events;

