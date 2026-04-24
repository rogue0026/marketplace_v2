CREATE TABLE IF NOT EXISTS payments (
                                        id BIGSERIAL PRIMARY KEY,
                                        order_id BIGINT REFERENCES orders(id),
                                        user_id BIGINT NOT NULL,
                                        total_price BIGINT NOT NULL,
                                        created_at TIMESTAMP NOT NULL DEFAULT now()
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

CREATE TABLE IF NOT EXISTS orders (
                                      id BIGSERIAL PRIMARY KEY,
                                      user_id BIGINT NOT NULL,
                                      total_price BIGINT NOT NULL,
                                      order_status VARCHAR(128),
);

CREATE TABLE IF NOT EXISTS orders_content (
                                              id BIGSERIAL PRIMARY KEY,
                                              order_id BIGINT REFERENCES orders (id) ON DELETE RESTRICT,
                                              user_id BIGINT NOT NULL,
                                              product_id BIGINT NOT NULL,
                                              product_quantity BIGINT NOT NULL,
                                              product_price_per_unit BIGINT NOT NULL
);