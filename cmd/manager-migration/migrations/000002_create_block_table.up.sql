CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS blocks
(
    id         uuid DEFAULT uuid_generate_v4(),
    time       TIMESTAMP WITH TIME ZONE NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE,

    hash        TEXT NOT NULL,
    height      DECIMAL(65, 0) NOT NULL,
    chain_id    VARCHAR(100) NOT NULL,
    
    header      JSONB NOT NULL,
    data        JSONB NOT NULL,
    evidence    JSONB NOT NULL,
    last_commit JSONB,
    numtxs      DECIMAL(65, 0) NOT NULL DEFAULT 0,

    PRIMARY KEY (id)
);


CREATE UNIQUE INDEX idx_blx_hash on blocks (chain_id, hash);
CREATE INDEX idx_blx_height on blocks (chain_id, height);
CREATE INDEX idx_blx_time on blocks (chain_id, time);
CREATE INDEX idx_partial_blx_height on blocks (height);
CREATE INDEX idx_partial_blx_numtxs_height on blocks (height, numtxs);
