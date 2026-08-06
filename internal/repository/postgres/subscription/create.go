package subRepo

import (
	"context"
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubRepository) Create(
	ctx context.Context,
	userID, productID uuid.UUID,
	productName string,
	expAt time.Time,
) (*domain.Subscription, error) {

	query := `
		INSERT INTO subscriptions (
			user_id,
			product_id,
			product_name,
			expires_at,
			status
		)
		VALUES (
			@user_id,
			@product_id,
			@product_name,
			@expires_at,
			@status
		)
		RETURNING
			id,
			user_id,
			product_id,
			product_name,
			expires_at,
			status,
			created_at,
			updated_at
	`

	args := pgx.NamedArgs{
		"user_id":      userID,
		"product_id":   productID,
		"product_name": productName,
		"expires_at":   expAt,
		"status":       domain.SubscriptionStatusActive,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return nil, mapSubscriptionPgError(err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[SubscriptionRow],
	)
	if err != nil {
		return nil, mapSubscriptionPgError(err)
	}

	sub, err := row.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("map subscription: %w", err)
	}

	return sub, nil
}
