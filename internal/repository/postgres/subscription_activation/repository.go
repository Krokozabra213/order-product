package subActiveRepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	pgRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	constraintSubscriptionActivationUnique = "idx_subscription_activations_unique"
)

type SubActiveRepo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *SubActiveRepo {
	return &SubActiveRepo{pool: pool}
}

func (r *SubActiveRepo) exec(
	ctx context.Context,
	query string,
	args ...any,
) (pgconn.CommandTag, error) {

	if tx, ok := infrapg.GetTx(ctx); ok {
		return tx.Exec(ctx, query, args...)
	}
	return r.pool.Exec(ctx, query, args...)
}

func mapSubscriptionActivationPgError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("subscription activation query: %w", err)
	}

	if pgErr.Code == pgRepo.UniqueViolation &&
		pgErr.ConstraintName == constraintSubscriptionActivationUnique {
		return domain.ErrAlreadyExists
	}

	return fmt.Errorf("subscription activation query: %w", err)
}
