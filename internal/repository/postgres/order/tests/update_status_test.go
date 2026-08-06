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

func TestOrderRepository_UpdateStatus_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	seedProduct(t, pool, product)

	order, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, order.ID, domain.OrderStatusCompleted)
	require.NoError(t, err)

	updated, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err)

	assert.Equal(t, domain.OrderStatusCompleted, updated.Status)
}

func TestOrderRepository_UpdateStatus_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	err := repo.UpdateStatus(
		context.Background(),
		uuid.New(),
		domain.OrderStatusCompleted,
	)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestOrderRepository_UpdateStatus_InvalidStatus(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	err := repo.UpdateStatus(
		context.Background(),
		uuid.New(),
		domain.OrderStatus("INVALID"),
	)

	require.Error(t, err)
}
