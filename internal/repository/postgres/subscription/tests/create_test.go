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

func TestSubscriptionRepository_Create_Success(t *testing.T) {
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

	expAt := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Microsecond)

	sub, err := repo.Create(
		ctx,
		userID,
		productID,
		"Test",
		expAt,
	)

	require.NoError(t, err)
	require.NotNil(t, sub)

	assert.Equal(t, userID, sub.UserID)
	assert.Equal(t, productID, sub.ProductID)
	assert.Equal(t, domain.SubscriptionStatusActive, sub.Status)
	assert.WithinDuration(t, expAt, sub.ExpiresAt, time.Second)
}

func TestSubscriptionRepository_Create_AlreadyExists(t *testing.T) {
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

	expAt := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Microsecond)

	_, err := repo.Create(ctx, userID, productID, "Test", expAt)
	require.NoError(t, err)

	sub, err := repo.Create(ctx, userID, productID, "Test", expAt)

	require.Error(t, err)
	require.Nil(t, sub)
	require.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestSubscriptionRepository_Create_WithUoW(t *testing.T) {
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

	expAt := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Microsecond)

	var sub *domain.Subscription

	err := uow.Do(ctx, func(txCtx context.Context) error {
		var err error
		sub, err = repo.Create(txCtx, userID, productID, "Test", expAt)
		return err
	})

	require.NoError(t, err)
	require.NotNil(t, sub)
}
