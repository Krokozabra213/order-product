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

func TestSubscriptionRepository_ListByUser_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	otherUserID := uuid.New()

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

	productID2 := uuid.New()
	seedProduct(t, pool, &domain.Product{
		ID:            productID2,
		Type:          domain.ProductTypeSubscription,
		Name:          "Test2",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	})

	seedSubscription(t, pool, userID, productID, domain.SubscriptionStatusActive)
	seedSubscription(t, pool, userID, productID2, domain.SubscriptionStatusActive)

	seedSubscription(t, pool, otherUserID, productID, domain.SubscriptionStatusActive)

	subs, err := repo.ListByUser(ctx, userID, domain.SubscriptionStatusActive)

	require.NoError(t, err)
	require.Len(t, subs, 2)

	for _, s := range subs {
		assert.Equal(t, userID, s.UserID)
		assert.Equal(t, domain.SubscriptionStatusActive, s.Status)
	}
}

func TestSubscriptionRepository_ListByUser_Empty(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subRepo.New(pool)

	subs, err := repo.ListByUser(
		context.Background(),
		uuid.New(),
		domain.SubscriptionStatusActive,
	)

	require.NoError(t, err)
	require.Empty(t, subs)
}

func TestSubscriptionRepository_ListByUser_WithUoW(t *testing.T) {
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

	seedSubscription(t, pool, userID, productID, domain.SubscriptionStatusActive)

	var subs []domain.Subscription

	err := uow.Do(ctx, func(txCtx context.Context) error {
		var err error
		subs, err = repo.ListByUser(txCtx, userID, domain.SubscriptionStatusActive)
		return err
	})

	require.NoError(t, err)
	require.Len(t, subs, 1)
}
