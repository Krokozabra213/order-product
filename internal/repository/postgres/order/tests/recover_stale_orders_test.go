//go:build integration

package tests

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	testutil "github.com/Krokozabra213/gargantua_common/pkg/testutils"
	"github.com/GargantuaLabs/order-product/internal/domain"
	orderRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/order"
	"github.com/GargantuaLabs/order-product/migrations"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderRepository_RecoverStaleOrders_Success(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Recover stale",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}
	seedProduct(t, pool, product)

	order, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, order.ID, domain.OrderStatusProcessing)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`UPDATE orders SET updated_at = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
		order.ID,
	)
	require.NoError(t, err)

	recovered, err := repo.RecoverStaleOrders(
		ctx,
		5*time.Minute,
		domain.OrderStatusProcessing,
		10,
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), recovered)

	updated, err := repo.GetByID(ctx, order.ID)
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusPending, updated.Status)
}

func TestOrderRepository_RecoverStaleOrders_SkipFresh(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	product := &domain.Product{
		ID:            uuid.New(),
		Type:          domain.ProductTypeSubscription,
		Name:          "Fresh",
		DurationDays:  30,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyTypeRub,
	}
	seedProduct(t, pool, product)

	order, err := repo.Create(ctx, userID, product)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, order.ID, domain.OrderStatusProcessing)
	require.NoError(t, err)

	recovered, err := repo.RecoverStaleOrders(
		ctx,
		5*time.Minute,
		domain.OrderStatusProcessing,
		10,
	)
	require.NoError(t, err)
	require.Equal(t, int64(0), recovered)
}

func TestOrderRepository_RecoverStaleOrders_Limit(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	total := 5

	for i := 0; i < total; i++ {
		p := &domain.Product{
			ID:            uuid.New(),
			Type:          domain.ProductTypeSubscription,
			Name:          fmt.Sprintf("Limit %d", i),
			DurationDays:  30,
			PriceAmount:   amount,
			PriceCurrency: domain.CurrencyTypeRub,
		}
		seedProduct(t, pool, p)

		order, err := repo.Create(ctx, userID, p)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, order.ID, domain.OrderStatusProcessing)
		require.NoError(t, err)

		_, err = pool.Exec(ctx,
			`UPDATE orders SET updated_at = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
			order.ID,
		)
		require.NoError(t, err)
	}

	recovered, err := repo.RecoverStaleOrders(
		ctx,
		5*time.Minute,
		domain.OrderStatusProcessing,
		3,
	)
	require.NoError(t, err)
	require.Equal(t, int64(3), recovered)

	var pendingCount int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE status = 'PENDING'`,
	).Scan(&pendingCount)
	require.NoError(t, err)

	require.Equal(t, 3, pendingCount)
}

func TestOrderRepository_RecoverStaleOrders_Concurrent(t *testing.T) {
	pool := testutil.SetupTestDB(t, migrations.Files)
	repo := orderRepo.New(pool)

	ctx := context.Background()
	userID := uuid.New()
	amount, _ := types.NewPositiveInt(int64(1000))

	total := 4

	for i := 0; i < total; i++ {
		p := &domain.Product{
			ID:            uuid.New(),
			Type:          domain.ProductTypeSubscription,
			Name:          fmt.Sprintf("Concurrent %d", i),
			DurationDays:  30,
			PriceAmount:   amount,
			PriceCurrency: domain.CurrencyTypeRub,
		}
		seedProduct(t, pool, p)

		order, err := repo.Create(ctx, userID, p)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, order.ID, domain.OrderStatusProcessing)
		require.NoError(t, err)

		_, err = pool.Exec(ctx,
			`UPDATE orders SET updated_at = NOW() - INTERVAL '10 minutes' WHERE id = $1`,
			order.ID,
		)
		require.NoError(t, err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var r1, r2 int64

	go func() {
		defer wg.Done()
		r1, _ = repo.RecoverStaleOrders(
			context.Background(),
			5*time.Minute,
			domain.OrderStatusProcessing,
			3,
		)
	}()

	go func() {
		defer wg.Done()
		r2, _ = repo.RecoverStaleOrders(
			context.Background(),
			5*time.Minute,
			domain.OrderStatusProcessing,
			3,
		)
	}()

	wg.Wait()

	totalRecovered := r1 + r2
	require.LessOrEqual(t, totalRecovered, int64(total))

	var pendingCount int
	err := pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE status = 'PENDING'`,
	).Scan(&pendingCount)
	require.NoError(t, err)

	require.Equal(t, int(totalRecovered), pendingCount)
}
