CREATE schema IF NOT EXISTS test;
CREATE TABLE IF NOT EXISTS test.customer
(
    id UUID PRIMARY KEY
    , first_name VARCHAR(255)
    , last_name VARCHAR(255)
    , created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW())::BIGINT)
    , updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW())::BIGINT)
    , deleted_at BIGINT
);