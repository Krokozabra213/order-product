package orderRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	pgRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *OrderRepository) Create(
	ctx context.Context,
	userID uuid.UUID,
	product *domain.Product,
) (*domain.Order, error) {

	query := `
		INSERT INTO orders (
			user_id,
			product_id, product_name, product_type,
			duration_days,
			amount, currency,
			status
		)
		VALUES (
			@user_id,
			@product_id, @product_name, @product_type,
			@duration_days,
			@amount, @currency,
			@status
		)
		RETURNING
			id,
			user_id,
			product_id,
			product_name,
			product_type,
			duration_days,
			amount,
			currency,
			status,
			balance_operation_id,
			paid_at,
			canceled_at,
			expired_at,
			created_at,
			updated_at
	`

	args := pgx.NamedArgs{
		"user_id":       userID,
		"product_id":    product.ID,
		"product_name":  product.Name,
		"product_type":  product.Type,
		"duration_days": product.DurationDays,
		"amount":        product.PriceAmount.Value(),
		"currency":      product.PriceCurrency,
		"status":        domain.OrderStatusPending,
	}

	rows, err := r.query(ctx, query, args)
	if err != nil {
		return nil, pgRepo.MapPgError(err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[OrderRow],
	)
	if err != nil {
		return nil, pgRepo.MapPgError(err)
	}

	order, err := row.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("map order to domain: %w", err)
	}

	return order, nil
}
