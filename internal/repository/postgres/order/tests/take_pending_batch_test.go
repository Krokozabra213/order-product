//go:build integration

package tests

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	orderRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/order"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_TakePendingBatch_Commit(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	var products []*domain.Product
	for i := 0; i < 3; i++ {
		p := &domain.Product{
			ID:            uuid.New(),
			Type:          domain.ProductTypeSubscription,
			Name:          fmt.Sprintf("Test %d", i),
			DurationDays:  30,
			PriceAmount:   amount,
			PriceCurrency: domain.CurrencyTypeRub,
		}
		seedProduct(t, pool, p)
		products = append(products, p)
	}

	for _, p := range products {
		_, err := repo.Create(ctx, userID, p)
		require.NoError(t, err)
	}

	var batch []domain.Order

	err := uow.Do(ctx, func(txCtx context.Context) error {
		var err error
		batch, err = repo.TakePendingBatch(txCtx, 2)
		return err
	})

	require.NoError(t, err)
	assert.Len(t, batch, 2)

	for _, o := range batch {
		order, err := repo.GetByID(ctx, o.ID)
		require.NoError(t, err)
		assert.Equal(t, domain.OrderStatusProcessing, order.Status)
	}
}

func TestOrderRepository_TakePendingBatch_Rollback(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Rollback test",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}

	seedProduct(t, pool, product)

	order, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)

	_ = uow.Do(ctx, func(txCtx context.Context) error {
		_, err := repo.TakePendingBatch(txCtx, 1)
		require.NoError(t, err)
		return errors.New("force rollback")
	})

	updated, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.OrderStatusPending, updated.Status)
}

func TestOrderRepository_TakePendingBatch_NoTx(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	_, err := repo.TakePendingBatch(context.Background(), 10)

	require.Error(t, err)
}

func TestOrderRepository_TakePendingBatch_Limit(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := orderRepo.New(pool)
	ctx := context.Background()

	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	for i := 0; i < 5; i++ {
		p := &domain.Product{
			ID:            uuid.New(),
			Type:          domain.ProductTypeSubscription,
			Name:          fmt.Sprintf("Limit %d", i),
			DurationDays:  30,
			PriceAmount:   amount,
			PriceCurrency: domain.CurrencyTypeRub,
		}
		seedProduct(t, pool, p)
		_, err := repo.Create(ctx, userID, p)
		require.NoError(t, err)
	}

	var batch []domain.Order
	err := uow.Do(ctx, func(txCtx context.Context) error {
		var err error
		batch, err = repo.TakePendingBatch(txCtx, 3)
		return err
	})

	require.NoError(t, err)
	assert.Len(t, batch, 3)
}

func TestOrderRepository_TakePendingBatch_Concurrent(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	uow := infrapg.NewUnitOfWork(pool)
	repo := orderRepo.New(pool)

	ctx := context.Background()

	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	totalOrders := 4

	for i := 0; i < totalOrders; i++ {
		p := &domain.Product{
			ID:            uuid.New(),
			Type:          domain.ProductTypeSubscription,
			Name:          fmt.Sprintf("Concurrent %d", i),
			DurationDays:  30,
			PriceAmount:   amount,
			PriceCurrency: domain.CurrencyTypeRub,
		}
		seedProduct(t, pool, p)

		_, err := repo.Create(ctx, userID, p)
		require.NoError(t, err)
	}

	var (
		batch1 []domain.Order
		batch2 []domain.Order
		wg     sync.WaitGroup
	)

	start := make(chan struct{})
	wg.Add(2)

	go func() {
		defer wg.Done()

		<-start

		_ = uow.Do(context.Background(), func(txCtx context.Context) error {
			var err error
			batch1, err = repo.TakePendingBatch(txCtx, 3)

			time.Sleep(200 * time.Millisecond)

			return err
		})
	}()

	go func() {
		defer wg.Done()

		<-start

		_ = uow.Do(context.Background(), func(txCtx context.Context) error {
			var err error
			batch2, err = repo.TakePendingBatch(txCtx, 3)
			return err
		})
	}()

	close(start)
	wg.Wait()

	totalTaken := len(batch1) + len(batch2)

	require.LessOrEqual(t, totalTaken, totalOrders)

	ids := make(map[uuid.UUID]struct{})

	for _, o := range batch1 {
		ids[o.ID] = struct{}{}
	}
	for _, o := range batch2 {
		if _, exists := ids[o.ID]; exists {
			t.Fatalf("duplicate order taken concurrently: %v", o.ID)
		}
		ids[o.ID] = struct{}{}
	}

	require.Equal(t, totalTaken, len(ids))

	for id := range ids {
		order, err := repo.GetByID(ctx, id)
		require.NoError(t, err)
		require.Equal(t, domain.OrderStatusProcessing, order.Status)
	}

	var pendingCount int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE status = 'PENDING'`,
	).Scan(&pendingCount)
	require.NoError(t, err)

	require.Equal(t, totalOrders-totalTaken, pendingCount)
}
