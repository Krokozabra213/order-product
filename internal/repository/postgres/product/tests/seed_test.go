//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func seedProduct(t *testing.T, pool *pgxpool.Pool, p domain.Product) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO products (
			id, type, name, description,
			duration_days, price_amount, price_currency,
			is_active, is_archived,
			created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
	`,
		p.ID, p.Type, p.Name, p.Description,
		p.DurationDays, p.PriceAmount.Value(), p.PriceCurrency,
		p.IsActive, p.IsArchived,
		p.CreatedAt, p.UpdatedAt,
	)

	require.NoError(t, err)
}

func seedComponent(t *testing.T, pool *pgxpool.Pool, parentID, childID uuid.UUID) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO product_components (
			parent_product_id, child_product_id
		) VALUES ($1, $2)
	`, parentID, childID)

	require.NoError(t, err)
}
