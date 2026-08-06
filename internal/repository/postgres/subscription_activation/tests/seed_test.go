//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func seedSubscription(
	t *testing.T,
	pool *pgxpool.Pool,
	userID, productID uuid.UUID,
	status domain.SubscriptionStatus,
) domain.Subscription {

	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	sub := domain.Subscription{
		ID:          uuid.New(),
		UserID:      userID,
		ProductID:   productID,
		ProductName: "Test Product",
		ExpiresAt:   now.Add(30 * 24 * time.Hour),
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := pool.Exec(context.Background(), `
		INSERT INTO subscriptions (
			id,
			user_id,
			product_id,
			product_name,
			expires_at,
			status,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
	`,
		sub.ID,
		sub.UserID,
		sub.ProductID,
		sub.ProductName,
		sub.ExpiresAt,
		sub.Status,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	require.NoError(t, err)

	return sub
}

func seedProduct(t *testing.T, pool *pgxpool.Pool, p *domain.Product) {
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

func seedOrder(
	t *testing.T,
	pool *pgxpool.Pool,
	orderID uuid.UUID,
	userID uuid.UUID,
	productID uuid.UUID,
) domain.Order {

	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	amount, _ := types.NewPositiveInt(int64(1000))
	order := domain.Order{
		ID:           orderID,
		UserID:       userID,
		ProductID:    productID,
		ProductName:  "Test Product",
		ProductType:  domain.ProductTypeSubscription,
		DurationDays: 30,
		Amount:       amount,
		Currency:     domain.CurrencyTypeRub,
		Status:       domain.OrderStatusCompleted,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err := pool.Exec(context.Background(), `
		INSERT INTO orders (
			id,
			user_id,
			product_id,
			product_name,
			product_type,
			duration_days,
			amount,
			currency,
			status,
			created_at,
			updated_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
		)
	`,
		order.ID,
		order.UserID,
		order.ProductID,
		order.ProductName,
		order.ProductType,
		order.DurationDays,
		order.Amount.Value(),
		order.Currency,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	)

	require.NoError(t, err)

	return order
}
