-- +goose Up
SELECT 'up SQL query';

CREATE TABLE subscription_activations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    subscription_id UUID NOT NULL REFERENCES subscriptions(id),
    order_id UUID NOT NULL REFERENCES orders(id),

    duration_days INT NOT NULL,
    started_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_subscription_activations_unique
ON subscription_activations(subscription_id, order_id);

-- +goose Down
SELECT 'down SQL query';

DROP TABLE IF EXISTS subscriptions;
