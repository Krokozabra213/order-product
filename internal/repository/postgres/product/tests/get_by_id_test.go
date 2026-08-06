//go:build integration

package tests

import (
	"context"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	productRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/product"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByID_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Microsecond)
	amount, _ := types.NewPositiveInt(int64(1000))

	product := domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Test product",
		Description:   "Description",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	seedProduct(t, pool, product)

	result, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, product.ID, result.ID)
	assert.Equal(t, product.Type, result.Type)
	assert.Equal(t, product.Name, result.Name)
	assert.Equal(t, product.Description, result.Description)
	assert.Equal(t, product.DurationDays, result.DurationDays)
	assert.Equal(t, product.PriceAmount, result.PriceAmount)
	assert.Equal(t, product.PriceCurrency, result.PriceCurrency)
	assert.Equal(t, product.IsActive, result.IsActive)
	assert.Equal(t, product.IsArchived, result.IsArchived)
	assert.WithinDuration(t, product.CreatedAt, result.CreatedAt, time.Second)
	assert.WithinDuration(t, product.UpdatedAt, result.UpdatedAt, time.Second)
}

func TestGetByID_NotFound(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)

	randomID := uuid.New()

	result, err := repo.GetByID(context.Background(), randomID)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
