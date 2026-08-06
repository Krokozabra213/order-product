package productRepo

import (
	"context"
	"fmt"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ProductRepository) ListWithComponents(
	ctx context.Context,
	onlyActive bool,
) ([]domain.ProductWithComponents, error) {

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
		WHERE (NOT @only_active OR p.is_active)
		  AND p.is_archived = FALSE
		ORDER BY p.created_at DESC
	`

	rows, err := r.query(ctx, query, pgx.NamedArgs{
		"only_active": onlyActive,
	})
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	rawRows, err := pgx.CollectRows(
		rows,
		pgx.RowToStructByName[productWithComponentRow],
	)
	if err != nil {
		return nil, fmt.Errorf("collect products: %w", err)
	}

	productsMap := make(map[uuid.UUID]*domain.ProductWithComponents)
	order := make([]uuid.UUID, 0)

	for _, r := range rawRows {

		existing, ok := productsMap[r.ID]
		if !ok {
			existing = &domain.ProductWithComponents{
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
			productsMap[r.ID] = existing
			order = append(order, r.ID)
		}

		if r.ChildID != nil && r.ChildName != nil {
			existing.Components = append(existing.Components,
				domain.BundleComponent{
					ProductID: *r.ChildID,
					Name:      *r.ChildName,
				},
			)
		}
	}

	result := make([]domain.ProductWithComponents, 0, len(order))
	for _, id := range order {
		result = append(result, *productsMap[id])
	}

	return result, nil
}
