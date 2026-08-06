//go:build integration

package tests

import (
	"context"
	"testing"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	orderRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/order"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_ListByUser_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))

	product1 := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	product2 := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	product3 := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	seedProduct(t, pool, product1)
	seedProduct(t, pool, product2)
	seedProduct(t, pool, product3)

	_, err := repo.Create(ctx, userID, product1)
	require.NoError(t, err)

	_, err = repo.Create(ctx, userID, product2)
	require.NoError(t, err)

	_, err = repo.Create(ctx, userID, product3)
	require.NoError(t, err)

	orders, err := repo.ListByUser(ctx, userID, 10, 0)
	require.NoError(t, err)

	assert.Len(t, orders, 3)

	for _, o := range orders {
		assert.Equal(t, userID, o.UserID)
	}
}

func TestOrderRepository_ListByUser_Empty(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	orders, err := repo.ListByUser(context.Background(), uuid.New(), 10, 0)

	require.NoError(t, err)
	assert.Empty(t, orders)
}
