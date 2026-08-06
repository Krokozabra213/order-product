//go:build integration

package tests

import (
	"context"
	"testing"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func seedProduct(t *testing.T, pool *pgxpool.Pool, p *domain.Product) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
		INSERT INTO products (
			id,
			type,
			name,
			duration_days,
			price_amount,
			price_currency,
			is_active,
			is_archived,
			created_at,
			updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,true,false,now(),now()
		)
	`,
		p.ID,
		p.Type,
		p.Name,
		p.DurationDays,
		p.PriceAmount.Value(),
		p.PriceCurrency,
	)

	require.NoError(t, err)
}
