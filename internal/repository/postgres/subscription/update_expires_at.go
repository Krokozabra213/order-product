package subRepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *SubRepository) UpdateExpiresAt(
	ctx context.Context,
	subID uuid.UUID,
	expAt time.Time,
) error {

	query := `
		UPDATE subscriptions
		SET
			expires_at = @expires_at
		WHERE id = @id
		RETURNING id
	`

	args := pgx.NamedArgs{
		"id":         subID,
		"expires_at": expAt,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("update expires_at: %w", err)
	}
	defer rows.Close()

	_, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("collect updated subscription: %w", err)
	}

	return nil
}
