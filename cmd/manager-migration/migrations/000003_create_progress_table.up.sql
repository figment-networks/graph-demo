CREATE TABLE IF NOT EXISTS progress
(
    chain_id    VARCHAR(100) NOT NULL,
    height      DECIMAL(65, 0) NOT NULL,

    PRIMARY KEY (chain_id)
);


CREATE UNIQUE INDEX idx_prog_chain on progress (chain_id);
