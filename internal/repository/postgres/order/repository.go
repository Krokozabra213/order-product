package orderRepo

import (
	"context"

	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) query(
	ctx context.Context,
	query string,
	args ...any,
) (pgx.Rows, error) {
	if tx, ok := infrapg.GetTx(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}
