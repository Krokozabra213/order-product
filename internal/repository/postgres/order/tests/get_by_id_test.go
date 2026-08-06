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

func TestOrderRepository_GetByID_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
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

	order, err := repo.Create(ctx, userID, &domain.Product{
		ID:            productID,
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	})
	require.NoError(t, err)

	result, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, order.ID, result.ID)
	assert.Equal(t, order.UserID, result.UserID)
}

func TestOrderRepository_GetByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	randomID := uuid.New()

	result, err := repo.GetByID(context.Background(), randomID)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
