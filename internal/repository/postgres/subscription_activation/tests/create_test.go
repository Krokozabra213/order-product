//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	subActiveRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/subscription_activation"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionActivationRepository_Create_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subActiveRepo.New(pool)

	ctx := context.Background()

	userID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

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

	seedOrder(t, pool, orderID, userID, productID)

	startedAt := time.Now().UTC().Truncate(time.Microsecond)
	expAt := startedAt.Add(30 * 24 * time.Hour)

	err := repo.Create(ctx, sub.ID, orderID, 30, startedAt, expAt)
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM subscription_activations WHERE subscription_id = $1`,
		sub.ID,
	).Scan(&count)

	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestSubscriptionActivationRepository_Create_AlreadyExists(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := subActiveRepo.New(pool)

	ctx := context.Background()

	userID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

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

	seedOrder(t, pool, orderID, userID, productID)

	startedAt := time.Now().UTC()
	expAt := startedAt.Add(30 * 24 * time.Hour)

	err := repo.Create(ctx, sub.ID, orderID, 30, startedAt, expAt)
	require.NoError(t, err)

	err = repo.Create(ctx, sub.ID, orderID, 30, startedAt, expAt)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestSubscriptionActivationRepository_Create_WithUoW(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := subActiveRepo.New(pool)

	ctx := context.Background()

	userID := uuid.New()
	productID := uuid.New()
	orderID := uuid.New()

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

	seedOrder(t, pool, orderID, userID, productID)

	err := uow.Do(ctx, func(txCtx context.Context) error {
		return repo.Create(
			txCtx,
			sub.ID,
			orderID,
			30,
			time.Now().UTC(),
			time.Now().UTC().Add(30*24*time.Hour),
		)
	})

	require.NoError(t, err)
}
