-- +goose Up

CREATE TABLE IF NOT EXISTS product_components (
    parent_product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    child_product_id  UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (parent_product_id, child_product_id),

    CONSTRAINT chk_no_self_reference CHECK (parent_product_id <> child_product_id)
);

CREATE INDEX IF NOT EXISTS idx_product_components_parent
    ON product_components(parent_product_id);

CREATE INDEX IF NOT EXISTS idx_product_components_child
    ON product_components(child_product_id);


-- +goose Down

DROP TABLE IF EXISTS product_components;