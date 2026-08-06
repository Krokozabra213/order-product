package subActiveRepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubActiveRepo) Create(
	ctx context.Context,
	subID, orderID uuid.UUID,
	duration int,
	startedAt time.Time,
	expAt time.Time,
) error {

	query := `
		INSERT INTO subscription_activations (
			subscription_id,
			order_id,
			duration_days,
			started_at,
			expires_at
		)
		VALUES (
			@subscription_id,
			@order_id,
			@duration_days,
			@started_at,
			@expires_at
		)
	`

	args := pgx.NamedArgs{
		"subscription_id": subID,
		"order_id":        orderID,
		"duration_days":   duration,
		"started_at":      startedAt,
		"expires_at":      expAt,
	}

	_, err := r.exec(ctx, query, args)
	if err != nil {
		return mapSubscriptionActivationPgError(err)
	}

	return nil
}
