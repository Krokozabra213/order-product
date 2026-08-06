package orderRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Order, error) {

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
		WHERE id = $1
	`

	rows, err := r.query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("get order by id: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[OrderRow],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect order: %w", err)
	}

	order, err := row.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("map order to domain: %w", err)
	}

	return order, nil
}
