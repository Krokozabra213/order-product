package orderRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) ListByUser(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int32,
) ([]domain.Order, error) {

	query := `
		SELECT
			id,
			user_id,
			product_id,
			product_name,
			product_type,
			duration_days,
			amount,
			currency,
			status,
			balance_operation_id,
			paid_at,
			canceled_at,
			expired_at,
			created_at,
			updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list orders by user: %w", err)
	}
	defer rows.Close()

	orderRows, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[OrderRow],
	)
	if err != nil {
		return nil, fmt.Errorf("collect orders: %w", err)
	}

	result := make([]domain.Order, 0, len(orderRows))

	for _, row := range orderRows {
		order, err := row.ToDomain()
		if err != nil {
			return nil, fmt.Errorf("map order to domain: %w", err)
		}
		result = append(result, *order)
	}

	return result, nil
}
