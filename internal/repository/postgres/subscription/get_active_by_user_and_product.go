package subRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubRepository) GetActiveByUserAndProduct(
	ctx context.Context,
	userID, productID uuid.UUID,
) (*domain.Subscription, error) {

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
		WHERE user_id = $1
		  AND product_id = $2
		  AND status = 'ACTIVE'
	`

	rows, err := r.query(ctx, query, userID, productID)
	if err != nil {
		return nil, fmt.Errorf("get active subscription: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[SubscriptionRow],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect subscription: %w", err)
	}

	sub, err := row.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("map subscription: %w", err)
	}

	return sub, nil
}
