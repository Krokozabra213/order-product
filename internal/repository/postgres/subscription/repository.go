package subRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	pgRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ConstraintActiveSubscription = "idx_one_active_subscription"
)

type SubRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *SubRepository {
	return &SubRepository{pool: pool}
}

func (r *SubRepository) query(
	ctx context.Context,
	query string,
	args ...any,
) (pgx.Rows, error) {
	if tx, ok := infrapg.GetTx(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return r.pool.Query(ctx, query, args...)
}

func mapSubscriptionPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgRepo.UniqueViolation &&
			pgErr.ConstraintName == ConstraintActiveSubscription {
			return domain.ErrAlreadyExists
		}
	}
	return fmt.Errorf("subscription query: %w", err)
}
