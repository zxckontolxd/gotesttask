-- вместо миграции
CREATE DATABASE wallet_db;

\connect wallet_db;

CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY,
    balance BIGINT NOT NULL
);

INSERT INTO wallets (id, balance) VALUES ('2b40e216-fb9b-4e9e-bfa8-0422da3b7be0', 1000);
INSERT INTO wallets (id, balance) VALUES ('2b40e216-fb9b-4e9e-bfa8-0422da3b7be1', 5000);