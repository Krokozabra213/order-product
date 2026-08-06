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

func TestGetBundleComponents_Empty(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)

	bundleID := uuid.New()

	result, err := repo.GetBundleComponents(context.Background(), bundleID)
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestGetBundleComponents_SingleComponent(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now()
	amount, _ := types.NewPositiveInt(int64(300))

	bundle := domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeBundle,
		Name:          "Bundle",
		DurationDays:  10,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	child := bundle
	child.ID = uuid.New()
	child.Type = domain.ProductTypeSubscription
	child.Name = "Child"

	seedProduct(t, pool, bundle)
	seedProduct(t, pool, child)
	seedComponent(t, pool, bundle.ID, child.ID)

	result, err := repo.GetBundleComponents(ctx, bundle.ID)
	require.NoError(t, err)

	require.Len(t, result, 1)
	assert.Equal(t, child.ID, result[0].ID)
}

func TestGetBundleComponents_MultipleComponents(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now()
	amount, _ := types.NewPositiveInt(int64(300))

	bundle := domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeBundle,
		Name:          "Bundle",
		DurationDays:  10,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	child1 := bundle
	child1.ID = uuid.New()
	child1.Type = domain.ProductTypeSubscription
	child1.Name = "Child 1"

	child2 := child1
	child2.ID = uuid.New()
	child2.Name = "Child 2"

	seedProduct(t, pool, bundle)
	seedProduct(t, pool, child1)
	seedProduct(t, pool, child2)

	seedComponent(t, pool, bundle.ID, child1.ID)
	seedComponent(t, pool, bundle.ID, child2.ID)

	result, err := repo.GetBundleComponents(ctx, bundle.ID)
	require.NoError(t, err)

	require.Len(t, result, 2)

	ids := []uuid.UUID{result[0].ID, result[1].ID}
	assert.Contains(t, ids, child1.ID)
	assert.Contains(t, ids, child2.ID)
}

func TestGetBundleComponents_IsolatedBetweenBundles(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now()
	amount, _ := types.NewPositiveInt(int64(300))

	bundle1 := domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeBundle,
		Name:          "Bundle 1",
		DurationDays:  10,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	bundle2 := bundle1
	bundle2.ID = uuid.New()
	bundle2.Name = "Bundle 2"

	child := bundle1
	child.ID = uuid.New()
	child.Type = domain.ProductTypeSubscription

	seedProduct(t, pool, bundle1)
	seedProduct(t, pool, bundle2)
	seedProduct(t, pool, child)

	seedComponent(t, pool, bundle1.ID, child.ID)

	result1, err := repo.GetBundleComponents(ctx, bundle1.ID)
	require.NoError(t, err)
	require.Len(t, result1, 1)

	result2, err := repo.GetBundleComponents(ctx, bundle2.ID)
	require.NoError(t, err)
	require.Empty(t, result2)
}
