package tests

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_ListProducts(t *testing.T) {

	type testCase struct {
		name       string
		onlyActive bool
		setup      func(deps *testDeps, onlyActive bool)

		expectErr      bool
		expectProducts bool
	}

	tests := []testCase{

		{
			name:       "cache hit",
			onlyActive: true,
			setup: func(deps *testDeps, onlyActive bool) {
				cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)

				expected := []domain.ProductWithComponents{
					{ID: uuid.New()},
				}

				deps.cache.EXPECT().
					GetProducts(gomock.Any(), cacheKey).
					Return(expected, nil)
			},
			expectErr:      false,
			expectProducts: true,
		},

		{
			name:       "cache miss db success",
			onlyActive: true,
			setup: func(deps *testDeps, onlyActive bool) {
				cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)

				expected := []domain.ProductWithComponents{
					{ID: uuid.New()},
				}

				deps.cache.EXPECT().
					GetProducts(gomock.Any(), cacheKey).
					Return(nil, nil)

				deps.cache.EXPECT().
					GetProducts(gomock.Any(), cacheKey).
					Return(nil, nil)

				deps.product.EXPECT().
					ListWithComponents(gomock.Any(), onlyActive).
					Return(expected, nil)

				deps.cache.EXPECT().
					SetProducts(gomock.Any(), cacheKey, expected, 20*time.Minute).
					Return(nil)
			},
			expectErr:      false,
			expectProducts: true,
		},

		{
			name:       "cache miss db error",
			onlyActive: false,
			setup: func(deps *testDeps, onlyActive bool) {
				cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)

				deps.cache.EXPECT().
					GetProducts(gomock.Any(), cacheKey).
					Return(nil, nil)

				deps.cache.EXPECT().
					GetProducts(gomock.Any(), cacheKey).
					Return(nil, nil)

				deps.product.EXPECT().
					ListWithComponents(gomock.Any(), onlyActive).
					Return(nil, errors.New("db error"))
			},
			expectErr:      true,
			expectProducts: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc, deps := newTestService(t)
			defer deps.ctrl.Finish()

			ctx := context.Background()

			tt.setup(deps, tt.onlyActive)

			products, err := svc.ListProducts(ctx, tt.onlyActive)

			if tt.expectErr {
				require.Error(t, err)
				require.Nil(t, products)
			} else {
				require.NoError(t, err)
				require.NotNil(t, products)
			}
		})
	}
}

func TestService_ListProducts_SingleflightConcurrent(t *testing.T) {

	svc, deps := newTestService(t)
	defer deps.ctrl.Finish()

	ctx := context.Background()
	onlyActive := true

	cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)

	expected := []domain.ProductWithComponents{
		{ID: uuid.New()},
	}

	deps.cache.EXPECT().
		GetProducts(gomock.Any(), cacheKey).
		Return(nil, nil).
		AnyTimes()

	deps.product.EXPECT().
		ListWithComponents(gomock.Any(), onlyActive).
		DoAndReturn(func(ctx context.Context, onlyActive bool) ([]domain.ProductWithComponents, error) {
			time.Sleep(50 * time.Millisecond)
			return expected, nil
		}).
		Times(1)

	deps.cache.EXPECT().
		SetProducts(gomock.Any(), cacheKey, expected, 20*time.Minute).
		Return(nil).
		Times(1)

	var wg sync.WaitGroup
	concurrency := 10

	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			products, err := svc.ListProducts(ctx, onlyActive)
			require.NoError(t, err)
			require.NotNil(t, products)
		}()
	}

	wg.Wait()
}

func TestService_ListProducts_CacheErrorFallback(t *testing.T) {

	svc, deps := newTestService(t)
	defer deps.ctrl.Finish()

	ctx := context.Background()
	onlyActive := true

	cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)

	expected := []domain.ProductWithComponents{
		{ID: uuid.New()},
	}

	deps.cache.EXPECT().
		GetProducts(gomock.Any(), cacheKey).
		Return(nil, errors.New("redis down"))

	deps.cache.EXPECT().
		GetProducts(gomock.Any(), cacheKey).
		Return(nil, errors.New("redis down"))

	deps.product.EXPECT().
		ListWithComponents(gomock.Any(), onlyActive).
		Return(expected, nil)

	deps.cache.EXPECT().
		SetProducts(gomock.Any(), cacheKey, expected, 20*time.Minute).
		Return(nil)

	products, err := svc.ListProducts(ctx, onlyActive)

	require.NoError(t, err)
	require.NotNil(t, products)
}
