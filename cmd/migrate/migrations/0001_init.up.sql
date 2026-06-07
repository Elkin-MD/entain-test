-- Idempotent schema: safe to run on every startup.

CREATE TABLE IF NOT EXISTS users (
    id      BIGINT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
    id             BIGSERIAL PRIMARY KEY,
    transaction_id TEXT        NOT NULL UNIQUE,
    user_id        BIGINT      NOT NULL REFERENCES users (id),
    state          TEXT        NOT NULL,
    source_type    TEXT        NOT NULL,
    amount         BIGINT      NOT NULL CHECK (amount >= 0),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions (user_id);

-- Predefined users 1, 2, 3 (balance in minor units / cents).
INSERT INTO users (id, balance) VALUES (1, 0), (2, 0), (3, 0)
ON CONFLICT (id) DO NOTHING;
