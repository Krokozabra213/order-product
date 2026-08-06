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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWithComponents_BasicScenarios(t *testing.T) {
	tests := []struct {
		name       string
		onlyActive bool
		seed       func(t *testing.T, pool *pgxpool.Pool)
		assert     func(t *testing.T, result []domain.ProductWithComponents)
	}{
		{
			name:       "empty db",
			onlyActive: false,
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				// ничего не сидим
			},
			assert: func(t *testing.T, result []domain.ProductWithComponents) {
				require.Empty(t, result)
			},
		},
		{
			name:       "product without components",
			onlyActive: false,
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				now := time.Now()
				amount, _ := types.NewPositiveInt(int64(300))

				product := domain.Product{
					ID:            uuid.New(),
					Type:          domain.ProductTypeSubscription,
					Name:          "Single",
					Description:   "No components",
					DurationDays:  10,
					PriceAmount:   amount,
					PriceCurrency: domain.CurrencyTypeRub,
					IsActive:      true,
					IsArchived:    false,
					CreatedAt:     now,
					UpdatedAt:     now,
				}

				seedProduct(t, pool, product)
			},
			assert: func(t *testing.T, result []domain.ProductWithComponents) {
				require.Len(t, result, 1)
				assert.Empty(t, result[0].Components)
			},
		},
		{
			name:       "only active filter",
			onlyActive: true,
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				now := time.Now()
				amount, _ := types.NewPositiveInt(int64(300))

				active := domain.Product{
					ID:            uuid.New(),
					Type:          domain.ProductTypeSubscription,
					Name:          "Active",
					DurationDays:  10,
					PriceAmount:   amount,
					PriceCurrency: domain.CurrencyTypeRub,
					IsActive:      true,
					IsArchived:    false,
					CreatedAt:     now,
					UpdatedAt:     now,
				}

				inactive := active
				inactive.ID = uuid.New()
				inactive.Name = "Inactive"
				inactive.IsActive = false

				seedProduct(t, pool, active)
				seedProduct(t, pool, inactive)
			},
			assert: func(t *testing.T, result []domain.ProductWithComponents) {
				require.Len(t, result, 1)
				assert.True(t, result[0].IsActive)
			},
		},
		{
			name:       "exclude archived",
			onlyActive: false,
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				now := time.Now()
				amount, _ := types.NewPositiveInt(int64(300))

				archived := domain.Product{
					ID:            uuid.New(),
					Type:          domain.ProductTypeSubscription,
					Name:          "Archived",
					DurationDays:  10,
					PriceAmount:   amount,
					PriceCurrency: domain.CurrencyTypeRub,
					IsActive:      true,
					IsArchived:    true,
					CreatedAt:     now,
					UpdatedAt:     now,
				}

				seedProduct(t, pool, archived)
			},
			assert: func(t *testing.T, result []domain.ProductWithComponents) {
				require.Empty(t, result)
			},
		},
		{
			name:       "order by created_at desc",
			onlyActive: false,
			seed: func(t *testing.T, pool *pgxpool.Pool) {
				now := time.Now()
				amount, _ := types.NewPositiveInt(int64(300))

				older := domain.Product{
					ID:            uuid.New(),
					Type:          domain.ProductTypeSubscription,
					Name:          "Old",
					DurationDays:  10,
					PriceAmount:   amount,
					PriceCurrency: domain.CurrencyTypeRub,
					IsActive:      true,
					IsArchived:    false,
					CreatedAt:     now.Add(-time.Hour),
					UpdatedAt:     now,
				}

				newer := older
				newer.ID = uuid.New()
				newer.Name = "New"
				newer.CreatedAt = now

				seedProduct(t, pool, older)
				seedProduct(t, pool, newer)
			},
			assert: func(t *testing.T, result []domain.ProductWithComponents) {
				require.Len(t, result, 2)
				assert.True(t, result[0].CreatedAt.After(result[1].CreatedAt) ||
					result[0].CreatedAt.Equal(result[1].CreatedAt))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := testutil.SetupTestDB(t, migrations.Files)
			repo := productRepo.New(pool)

			tt.seed(t, pool)

			result, err := repo.ListWithComponents(context.Background(), tt.onlyActive)
			require.NoError(t, err)

			tt.assert(t, result)
		})
	}
}

func TestListWithComponents_WithComponents(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now()

	parentID := uuid.New()
	childID1 := uuid.New()
	childID2 := uuid.New()

	amount, _ := types.NewPositiveInt(int64(1000))

	parent := domain.Product{
		ID:            parentID,
		Type:          domain.ProductTypeBundle,
		Name:          "Bundle",
		Description:   "Bundle desc",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	amount2, _ := types.NewPositiveInt(int64(300))

	child1 := domain.Product{
		ID:            childID1,
		Type:          domain.ProductTypeSubscription,
		Name:          "Child 1",
		Description:   "Child 1 desc",
		DurationDays:  10,
		PriceAmount:   amount2,
		PriceCurrency: domain.CurrencyTypeRub,
		IsActive:      true,
		IsArchived:    false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	child2 := child1
	child2.ID = childID2
	child2.Name = "Child 2"

	seedProduct(t, pool, parent)
	seedProduct(t, pool, child1)
	seedProduct(t, pool, child2)

	seedComponent(t, pool, parentID, childID1)
	seedComponent(t, pool, parentID, childID2)

	result, err := repo.ListWithComponents(ctx, false)
	require.NoError(t, err)

	require.Len(t, result, 3)

	var bundle domain.ProductWithComponents
	for _, p := range result {
		if p.ID == parentID {
			bundle = p
		}
	}

	require.Len(t, bundle.Components, 2)

	componentIDs := []uuid.UUID{
		bundle.Components[0].ProductID,
		bundle.Components[1].ProductID,
	}

	assert.Contains(t, componentIDs, childID1)
	assert.Contains(t, componentIDs, childID2)
}

func TestListWithComponents_MultipleBundles(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := productRepo.New(pool)
	ctx := context.Background()

	now := time.Now()

	amount, _ := types.NewPositiveInt(int64(300))

	bundle1 := domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeBundle,
		Name:          "Bundle 1",
		Description:   "No components",
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

	child1 := bundle1
	child1.ID = uuid.New()
	child1.Type = domain.ProductTypeSubscription
	child1.Name = "Child 1"

	child2 := child1
	child2.ID = uuid.New()
	child2.Name = "Child 2"

	seedProduct(t, pool, bundle1)
	seedProduct(t, pool, bundle2)
	seedProduct(t, pool, child1)
	seedProduct(t, pool, child2)

	seedComponent(t, pool, bundle1.ID, child1.ID)
	seedComponent(t, pool, bundle2.ID, child2.ID)

	result, err := repo.ListWithComponents(ctx, false)
	require.NoError(t, err)

	var b1, b2 domain.ProductWithComponents

	for _, p := range result {
		if p.ID == bundle1.ID {
			b1 = p
		}
		if p.ID == bundle2.ID {
			b2 = p
		}
	}

	require.Len(t, b1.Components, 1)
	require.Len(t, b2.Components, 1)

	assert.Equal(t, child1.ID, b1.Components[0].ProductID)
	assert.Equal(t, child2.ID, b2.Components[0].ProductID)
}
