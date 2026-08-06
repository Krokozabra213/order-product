package subRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	status domain.SubscriptionStatus,
) ([]domain.Subscription, error) {

	query := `
		SELECT
			id,
			user_id,
			product_id,
			product_name,
			expires_at,
			status,
			created_at,
			updated_at
		FROM subscriptions
		WHERE user_id = @user_id
		  AND status = @status
		ORDER BY created_at DESC
	`

	args := pgx.NamedArgs{
		"user_id": userID,
		"status":  status,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions by user: %w", err)
	}
	defer rows.Close()

	subRows, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[SubscriptionRow],
	)
	if err != nil {
		return nil, fmt.Errorf("collect subscriptions: %w", err)
	}

	result := make([]domain.Subscription, 0, len(subRows))

	for _, row := range subRows {
		sub, err := row.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("map subscription: %w", err)
		}
		result = append(result, *sub)
	}

	return result, nil
}
