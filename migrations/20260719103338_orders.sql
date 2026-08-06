-- +goose Up
SELECT 'up SQL query';

CREATE TABLE orders (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id           UUID NOT NULL,

    product_id        UUID NOT NULL REFERENCES products(id),
    product_name      VARCHAR(255) NOT NULL,
    product_type      product_type NOT NULL,
    duration_days     INT NOT NULL,

    amount            BIGINT NOT NULL,
    currency          VARCHAR(10) NOT NULL DEFAULT 'RUB',

    status            VARCHAR(50) NOT NULL DEFAULT 'PENDING',

    balance_operation_id        UUID,

    paid_at           TIMESTAMP,
    canceled_at       TIMESTAMP,
    expired_at        TIMESTAMP,

    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),

    CHECK (currency IN ('RUB')),
    CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED', 'REFUNDED', 'EXPIRED'))
);

CREATE INDEX idx_orders_user_created_at ON orders(user_id, created_at DESC);

CREATE UNIQUE INDEX idx_orders_user_product_pending
ON orders(user_id, product_id)
WHERE status = 'PENDING';

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS orders;