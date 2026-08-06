package orderRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.OrderStatus,
) error {

	isCompleted := status == domain.OrderStatusCompleted

	query := `
		UPDATE orders
		SET
			status = @status,
			paid_at = CASE
				WHEN @is_completed AND paid_at IS NULL THEN NOW()
				ELSE paid_at
			END,
			updated_at = NOW()
		WHERE id = @id
		RETURNING id
	`

	args := pgx.NamedArgs{
		"status":       status,
		"id":           id,
		"is_completed": isCompleted,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	defer rows.Close()

	_, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("collect updated order: %w", err)
	}

	return nil
}
