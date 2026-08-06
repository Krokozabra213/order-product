//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	orderRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/order"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_Create_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1500))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Pro Plan",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	seedProduct(t, pool, product)

	order, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)
	require.NotNil(t, order)

	assert.NotEqual(t, uuid.Nil, order.ID)

	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, product.ID, order.ProductID)
	assert.Equal(t, product.Name, order.ProductName)
	assert.Equal(t, product.Type, order.ProductType)
	assert.Equal(t, product.DurationDays, order.DurationDays)
	assert.Equal(t, amount, order.Amount)
	assert.Equal(t, product.PriceCurrency, order.Currency)
	assert.Equal(t, domain.OrderStatusPending, order.Status)

	assert.WithinDuration(t, time.Now(), order.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), order.UpdatedAt, time.Second)

	var exists bool
	err = pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM orders WHERE id = $1)`,
		order.ID,
	).Scan(&exists)

	require.NoError(t, err)
	assert.True(t, exists)

	assert.Nil(t, order.BalanceOperationID)
	assert.Nil(t, order.PaidAt)
	assert.Nil(t, order.CancelledAt)
	assert.Nil(t, order.ExpiredAt)
}

func TestOrderRepository_Create_FKViolation(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(), // не существует в БД
		Type:          domain.ProductTypeSubscription,
		Name:          "Ghost",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	order, err := repo.Create(ctx, userID, product)

	require.Error(t, err)
	assert.Nil(t, order)
}

func TestOrderRepository_Create_AlreadyExists(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Pro",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	seedProduct(t, pool, product)

	_, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)

	order, err := repo.Create(ctx, userID, product)

	require.Error(t, err)
	assert.Nil(t, order)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}
