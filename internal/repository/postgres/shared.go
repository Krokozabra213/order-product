package pgRepo

import (
	"errors"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	UniqueViolation = "23505"
)

func MapPgError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case UniqueViolation:
			return domain.ErrAlreadyExists
		}
	}
	return err
}
