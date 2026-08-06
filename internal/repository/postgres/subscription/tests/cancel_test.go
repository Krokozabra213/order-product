//go:build integration

package tests

import (
	"context"
	"testing"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	subRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/subscription"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRepository_Cancel_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))
	seedProduct(t, pool, &domain.Product{
		ID:            productID,
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	})

	sub := seedSubscription(
		t,
		pool,
		userID,
		productID,
		domain.SubscriptionStatusActive,
	)

	err := repo.Cancel(ctx, sub.ID)
	require.NoError(t, err)

	var status string
	err = pool.QueryRow(ctx,
		`SELECT status FROM subscriptions WHERE id = $1`,
		sub.ID,
	).Scan(&status)

	require.NoError(t, err)
	require.Equal(t, "CANCELED", status)
}

func TestSubscriptionRepository_Cancel_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	err := repo.Cancel(context.Background(), uuid.New())

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestSubscriptionRepository_Cancel_Idempotent(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	productID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))
	seedProduct(t, pool, &domain.Product{
		ID:            productID,
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	})

	sub := seedSubscription(
		t,
		pool,
		userID,
		productID,
		domain.SubscriptionStatusActive,
	)

	err := repo.Cancel(ctx, sub.ID)
	require.NoError(t, err)

	err = repo.Cancel(ctx, sub.ID)
	require.NoError(t, err)

	var status string
	err = pool.QueryRow(ctx,
		`SELECT status FROM subscriptions WHERE id = $1`,
		sub.ID,
	).Scan(&status)

	require.NoError(t, err)
	require.Equal(t, "CANCELED", status)
}
