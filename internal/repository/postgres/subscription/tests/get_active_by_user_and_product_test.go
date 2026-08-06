//go:build integration

package tests

import (
	"context"
	"testing"

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

func TestSubscriptionRepository_GetActiveByUserAndProduct_Success(t *testing.T) {
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

	activeSub := seedSubscription(
		t,
		pool,
		userID,
		productID,
		domain.SubscriptionStatusActive,
	)

	var result *domain.Subscription

	err := uow.Do(ctx, func(txCtx context.Context) error {
		var err error
		result, err = repo.GetActiveByUserAndProduct(txCtx, userID, productID)
		return err
	})

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, activeSub.ID, result.ID)
	assert.Equal(t, domain.SubscriptionStatusActive, result.Status)
}

func TestSubscriptionRepository_GetActiveByUserAndProduct_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := subRepo.New(pool)

	err := uow.Do(context.Background(), func(txCtx context.Context) error {
		sub, err := repo.GetActiveByUserAndProduct(
			txCtx,
			uuid.New(),
			uuid.New(),
		)

		require.Error(t, err)
		require.Nil(t, sub)
		require.ErrorIs(t, err, domain.ErrNotFound)

		return nil
	})

	require.NoError(t, err)
}
