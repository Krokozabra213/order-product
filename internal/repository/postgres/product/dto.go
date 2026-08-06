package productRepo

import (
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
)

type ProductRow struct {
	ID            uuid.UUID `db:"id"`
	Type          string    `db:"type"`
	Name          string    `db:"name"`
	Description   *string   `db:"description,omitempty"`
	DurationDays  int       `db:"duration_days"`
	PriceAmount   int64     `db:"price_amount"`
	PriceCurrency string    `db:"price_currency"`
	IsActive      bool      `db:"is_active"`
	IsArchived    bool      `db:"is_archived"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

func (row ProductRow) toDomain() domain.Product {
	amount, _ := types.NewPositiveInt(row.PriceAmount)
	return domain.Product{
		ID:            row.ID,
		Type:          domain.ProductType(row.Type),
		Name:          row.Name,
		Description:   derefString(row.Description),
		DurationDays:  row.DurationDays,
		PriceAmount:   amount,
		PriceCurrency: domain.CurrencyType(row.PriceCurrency),
		IsActive:      row.IsActive,
		IsArchived:    row.IsArchived,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

type productWithComponentRow struct {
	ID            uuid.UUID          `db:"id"`
	Type          domain.ProductType `db:"type"`
	Name          string             `db:"name"`
	Description   string             `db:"description"`
	DurationDays  int                `db:"duration_days"`
	PriceAmount   int64              `db:"price_amount"`
	PriceCurrency string             `db:"price_currency"`
	IsActive      bool               `db:"is_active"`
	IsArchived    bool               `db:"is_archived"`
	CreatedAt     time.Time          `db:"created_at"`
	UpdatedAt     time.Time          `db:"updated_at"`
	ChildID       *uuid.UUID         `db:"child_id"`
	ChildName     *string            `db:"child_name"`
}
