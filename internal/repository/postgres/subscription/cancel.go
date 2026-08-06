package subRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubRepository) Cancel(
	ctx context.Context,
	subscriptionID uuid.UUID,
) error {

	query := `
		UPDATE subscriptions
		SET
			status = @status
		WHERE id = @id
		RETURNING id
	`

	args := pgx.NamedArgs{
		"id":     subscriptionID,
		"status": domain.SubscriptionStatusCanceled,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}
	defer rows.Close()

	_, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("collect canceled subscription: %w", err)
	}

	return nil
}
