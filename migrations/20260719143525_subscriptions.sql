-- +goose Up
SELECT 'up SQL query';

CREATE TABLE subscriptions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id           UUID NOT NULL,
    product_id        UUID NOT NULL REFERENCES products(id),
    product_name      VARCHAR(255) NOT NULL,

    expires_at        TIMESTAMP NOT NULL,

    status            VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',

    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),

    CHECK (status IN ('ACTIVE', 'EXPIRED', 'CANCELED'))
);

CREATE UNIQUE INDEX idx_one_active_subscription
ON subscriptions(user_id, product_id)
WHERE status = 'ACTIVE';

CREATE INDEX idx_subscriptions_user_product
ON subscriptions(user_id, product_id);

CREATE INDEX idx_subscriptions_user_status
    ON subscriptions(user_id, status);

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS subscriptions;