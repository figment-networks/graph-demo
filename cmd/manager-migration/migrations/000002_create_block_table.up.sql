CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS blocks
(
    chain_id    VARCHAR(100) NOT NULL,
    height      DECIMAL(65, 0) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,

    hash        TEXT NOT NULL,
    time       TIMESTAMP WITH TIME ZONE NOT NULL,

    header      JSONB NOT NULL,
    data        JSONB NOT NULL,
    evidence    JSONB,
    last_commit JSONB,
);


CREATE UNIQUE INDEX idx_blx_hash on blocks (chain_id, hash);
CREATE INDEX idx_blx_height on blocks (chain_id, height);
CREATE INDEX idx_blx_time on blocks (chain_id, time);
