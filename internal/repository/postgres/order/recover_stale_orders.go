package orderRepo

import (
	"context"
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) RecoverStaleOrders(
	ctx context.Context,
	timeout time.Duration,
	status domain.OrderStatus,
	limit int,
) (int64, error) {

	query := `
		UPDATE orders
		SET status = @new_status,
		    updated_at = NOW()
		WHERE id IN (
			SELECT id
			FROM orders
			WHERE status = @current_status
			  AND updated_at < NOW() - (@timeout_seconds * INTERVAL '1 second')
			ORDER BY updated_at
			FOR UPDATE SKIP LOCKED
			LIMIT @limit
		)
	`

	args := pgx.NamedArgs{
		"new_status":      domain.OrderStatusPending,
		"current_status":  status,
		"timeout_seconds": int(timeout.Seconds()),
		"limit":           limit,
	}

	result, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return 0, fmt.Errorf("recover stale orders: %w", err)
	}

	return result.RowsAffected(), nil
}
