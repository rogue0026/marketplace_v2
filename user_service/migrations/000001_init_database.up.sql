CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     username VARCHAR(64) NOT NULL,
                                     password_hash VARCHAR(512) NOT NULL,
                                     CONSTRAINT username_unique UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS wallets (
                                       id BIGSERIAL PRIMARY KEY,
                                       user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
                                       balance BIGINT NOT NULL DEFAULT 0,
                                       CONSTRAINT balance_is_positive CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS basket (
                                      id BIGSERIAL PRIMARY KEY,
                                      user_id BIGINT REFERENCES users (id) ON DELETE CASCADE,
                                      product_id BIGINT NOT NULL,
                                      product_quantity BIGINT NOT NULL CHECK (product_quantity > 0) DEFAULT 1,
                                      CONSTRAINT user_id_product_id_unique UNIQUE (user_id, product_id)
);

CREATE TABLE IF NOT EXISTS outbox (
                                      id BIGSERIAL PRIMARY KEY,
                                      topic_name TEXT NOT NULL,
                                      partition_key TEXT NOT NULL,
                                      payload JSONB NOT NULL,
                                      status VARCHAR(64) NOT NULL DEFAULT 'PENDING',
                                      attempts INT NOT NULL DEFAULT 0,
                                      last_error TEXT,
                                      created_at TIMESTAMP NOT NULL DEFAULT now(),
                                      published_at TIMESTAMP,
                                      CONSTRAINT status_one_of CHECK ( status in ('PENDING', 'PROCESSED') )
);