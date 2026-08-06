package productRepo

import (
	"context"

	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) query(
	ctx context.Context,
	query string,
	args ...any,
) (pgx.Rows, error) {
	if tx, ok := infrapg.GetTx(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}
