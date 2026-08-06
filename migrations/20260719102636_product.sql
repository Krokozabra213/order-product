-- +goose Up
SELECT 'up SQL query';

CREATE TYPE product_type AS ENUM (
    'SUBSCRIPTION',
    'BUNDLE'
);

CREATE TABLE products (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    type              product_type NOT NULL,

    name              VARCHAR(255) NOT NULL,
    description       TEXT,

    duration_days     INT NOT NULL,

    price_amount      BIGINT NOT NULL,
    price_currency    VARCHAR(10) NOT NULL DEFAULT 'RUB',

    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    is_archived       BOOLEAN NOT NULL DEFAULT FALSE,

    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMP NOT NULL DEFAULT NOW(),

    CHECK (price_currency IN ('RUB')),
    CHECK (price_amount > 0)
);

CREATE INDEX idx_products_created_at ON products(created_at DESC);

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS products;
