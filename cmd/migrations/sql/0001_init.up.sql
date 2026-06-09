CREATE TABLE IF NOT EXISTS users (
    id      BIGINT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transaction_states (
    id   SMALLINT    PRIMARY KEY,
    name VARCHAR(16) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS source_types (
    id   SMALLINT    PRIMARY KEY,
    name VARCHAR(16) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS transactions (
    id             BIGSERIAL    PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL UNIQUE,
    user_id        BIGINT       NOT NULL REFERENCES users (id),
    state          SMALLINT     NOT NULL REFERENCES transaction_states (id),
    source_type    SMALLINT     NOT NULL REFERENCES source_types (id),
    amount         BIGINT       NOT NULL CHECK (amount >= 0),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);

INSERT INTO transaction_states (id, name) VALUES (1, 'win'), (2, 'lose')
ON CONFLICT (id) DO NOTHING;

INSERT INTO source_types (id, name) VALUES (1, 'game'), (2, 'server'), (3, 'payment')
ON CONFLICT (id) DO NOTHING;

INSERT INTO users (id, balance) VALUES (1, 0), (2, 0), (3, 0)
ON CONFLICT (id) DO NOTHING;
