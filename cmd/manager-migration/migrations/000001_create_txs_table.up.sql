CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS transactions
(
    id         uuid DEFAULT uuid_generate_v4(),
    

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,

    chain_id        VARCHAR(100) NOT NULL,
    height          DECIMAL(65, 0) NOT NULL,
    hash            TEXT NOT NULL,
    block_hash      TEXT NOT NULL,
    time       TIMESTAMP WITH TIME ZONE NOT NULL,
    
    code_space  TEXT,
    code        INT,
    gas_wanted  DECIMAL(65, 0),
    gas_used    DECIMAL(65, 0),
    info        TEXT,
    memo        TEXT,
    result      TEXT,
    signatures  TEXT[],

    auth_info                       JSONB,  
    extension_options               JSONB,
    logs                            JSONB,
    messages                        JSONB,
    non_critical_extension_options  JSONB,

    raw_log     BYTEA NOT NULL,
    tx_raw      JSONB,

    PRIMARY KEY (id)
);


CREATE UNIQUE INDEX idx_tx_ev_hash on transactions (chain_id, hash, height);
CREATE INDEX idx_tx_ev_ch_height on transactions (chain_id, height);
CREATE INDEX idx_tx_ev_ch_time on transactions (chain_id, time);
CREATE INDEX idx_tx_ev_block_hash on transactions (chain_id, block_hash);

