//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	subRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/subscription"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionRepository_UpdateExpiresAt_Success(t *testing.T) {
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

	newExp := time.Now().UTC().
		Add(90 * 24 * time.Hour).
		Truncate(time.Microsecond)

	err := repo.UpdateExpiresAt(ctx, sub.ID, newExp)
	require.NoError(t, err)

	var updatedExp time.Time
	err = pool.QueryRow(ctx,
		`SELECT expires_at FROM subscriptions WHERE id = $1`,
		sub.ID,
	).Scan(&updatedExp)
	require.NoError(t, err)

	assert.WithinDuration(t,
		newExp,
		updatedExp,
		time.Second,
	)
}

func TestSubscriptionRepository_UpdateExpiresAt_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	err := repo.UpdateExpiresAt(
		context.Background(),
		uuid.New(),
		time.Now().UTC(),
	)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrNotFound)
}

func TestSubscriptionRepository_UpdateExpiresAt_WithUoW(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
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

	newExp := time.Now().UTC().
		Add(30 * 24 * time.Hour).
		Truncate(time.Microsecond)

	err := uow.Do(ctx, func(txCtx context.Context) error {
		return repo.UpdateExpiresAt(txCtx, sub.ID, newExp)
	})

	require.NoError(t, err)
}
