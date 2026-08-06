package productRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Product, error) {

	query := `
		SELECT id, type, name, description,
		       duration_days, price_amount, price_currency,
		       is_active, is_archived,
		       created_at, updated_at
		FROM products
		WHERE id = @id
	`

	rows, err := r.query(ctx, query, pgx.NamedArgs{
		"id": id,
	})
	if err != nil {
		return nil, fmt.Errorf("get product by id: %w", err)
	}
	defer rows.Close()

	rowSelect, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[ProductRow],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("collect product by id: %w", err)
	}

	product := rowSelect.toDomain()

	return &product, nil
}
