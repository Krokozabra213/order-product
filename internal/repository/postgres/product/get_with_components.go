package productRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ProductRepository) GetWithComponents(
	ctx context.Context,
	productID uuid.UUID,
) (*domain.ProductWithComponents, error) {

	query := `
		SELECT 
			p.id,
			p.type,
			p.name,
			p.description,
			p.duration_days,
			p.price_amount,
			p.price_currency,
			p.is_active,
			p.is_archived,
			p.created_at,
			p.updated_at,
			pc.child_product_id AS child_id,
			cp.name AS child_name
		FROM products p
		LEFT JOIN product_components pc 
			ON pc.parent_product_id = p.id
		LEFT JOIN products cp 
			ON cp.id = pc.child_product_id
		WHERE p.id = @product_id
		  AND p.is_archived = FALSE
	`

	rows, err := r.query(ctx, query, pgx.NamedArgs{
		"product_id": productID,
	})
	if err != nil {
		return nil, fmt.Errorf("query product: %w", err)
	}
	defer rows.Close()

	rawRows, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[productWithComponentRow],
	)
	if err != nil {
		return nil, fmt.Errorf("collect product: %w", err)
	}

	if len(rawRows) == 0 {
		return nil, domain.ErrNotFound
	}

	var result *domain.ProductWithComponents

	for _, r := range rawRows {

		if result == nil {
			result = &domain.ProductWithComponents{
				ID:            r.ID,
				Type:          r.Type,
				Name:          r.Name,
				Description:   r.Description,
				DurationDays:  r.DurationDays,
				PriceAmount:   r.PriceAmount,
				PriceCurrency: r.PriceCurrency,
				IsActive:      r.IsActive,
				IsArchived:    r.IsArchived,
				CreatedAt:     r.CreatedAt,
				UpdatedAt:     r.UpdatedAt,
			}
		}

		if r.ChildID != nil && r.ChildName != nil {
			result.Components = append(result.Components,
				domain.BundleComponent{
					ProductID: *r.ChildID,
					Name:      *r.ChildName,
				},
			)
		}
	}

	return result, nil
}
