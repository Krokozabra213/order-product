package productRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ProductRepository) GetBundleComponents(
	ctx context.Context,
	bundleID uuid.UUID,
) ([]domain.Product, error) {

	query := `
		SELECT p.id, p.type, p.name, p.description,
		       p.duration_days, p.price_amount, p.price_currency,
		       p.is_active, p.is_archived,
		       p.created_at, p.updated_at
		FROM product_components pc
		JOIN products p ON p.id = pc.child_product_id
		WHERE pc.parent_product_id = @bundle_id
	`

	rows, err := r.query(ctx, query, pgx.NamedArgs{
		"bundle_id": bundleID,
	})
	if err != nil {
		return nil, fmt.Errorf("query bundle components: %w", err)
	}
	defer rows.Close()

	productRows, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[ProductRow],
	)
	if err != nil {
		return nil, fmt.Errorf("collect bundle components: %w", err)
	}

	products := make([]domain.Product, len(productRows))
	for i, row := range productRows {
		products[i] = row.toDomain()
	}

	return products, nil
}
