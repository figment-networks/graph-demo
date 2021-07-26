CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS transactions
(
    id         uuid DEFAULT uuid_generate_v4(),
    time       TIMESTAMP WITH TIME ZONE NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,

    chain_id        VARCHAR(100) NOT NULL,

    height          DECIMAL(65, 0) NOT NULL,
    hash            TEXT NOT NULL,
    block_hash      TEXT NOT NULL,

    type        VARCHAR(100)[] NOT NULL,
    parties     TEXT[] NOT NULL,
    senders     TEXT[] NOT NULL,
    recipients  TEXT[] NOT NULL,

    fee         JSONB NOT NULL,
    gas_wanted  DECIMAL(65, 0) NOT NULL,
    gas_used    DECIMAL(65, 0) NOT NULL,

    data    JSONB NOT NULL,
    raw     BYTEA NOT NULL,
    raw_log BYTEA NOT NULL,
    memo    TEXT NOT NULL,

    has_error BOOLEAN,

    PRIMARY KEY (id)
);


CREATE UNIQUE INDEX idx_tx_ev_hash on transactions (chain_id, hash, height);
CREATE INDEX idx_tx_ev_ch_height on transactions (chain_id, height);
CREATE INDEX idx_tx_ev_ch_time on transactions (chain_id, time);
CREATE INDEX idx_tx_ev_block_hash on transactions (chain_id, block_hash);
CREATE INDEX idx_tx_ev_parties_gin ON transactions USING GIN(parties);
CREATE INDEX idx_partial_tx_ev_height on transactions (height);
CREATE INDEX transaction_types_idx ON transactions USING GIN (type);
CREATE INDEX transaction_senders_idx ON transactions USING GIN (senders);
CREATE INDEX transaction_recipients_idx ON transactions USING GIN (recipients);

