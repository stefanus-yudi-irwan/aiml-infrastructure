CREATE TABLE IF NOT EXISTS customer
(
    id CHAR(36) PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
    updated_at BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
    deleted_at BIGINT
);