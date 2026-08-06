package orderRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) TakePendingBatch(
	ctx context.Context,
	limit int,
) ([]domain.Order, error) {

	tx, ok := infrapg.GetTx(ctx)
	if !ok {
		return nil, domain.ErrNoTx
	}

	query := `
		WITH cte AS (
			SELECT id
			FROM orders
			WHERE status = $1
			ORDER BY created_at
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		)
		UPDATE orders o
		SET
			status = $3
		FROM cte
		WHERE o.id = cte.id
		RETURNING
			o.id,
			o.user_id,
			o.product_id,
			o.product_name,
			o.product_type,
			o.duration_days,
			o.amount,
			o.currency,
			o.status,
			o.balance_operation_id,
			o.paid_at,
			o.canceled_at,
			o.expired_at,
			o.created_at,
			o.updated_at
	`

	rows, err := tx.Query(ctx, query,
		domain.OrderStatusPending,
		limit,
		domain.OrderStatusProcessing,
	)
	if err != nil {
		return nil, fmt.Errorf("take pending batch: %w", err)
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
			return nil, fmt.Errorf("map order: %w", err)
		}
		result = append(result, *order)
	}

	return result, nil
}
