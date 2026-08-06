package domain

import (
	"time"

	"github.com/google/uuid"
)

type BundleComponent struct {
	ProductID uuid.UUID
	Name      string
}

type ProductWithComponents struct {
	ID   uuid.UUID
	Type ProductType

	Name        string
	Description string

	DurationDays int

	PriceAmount   int64
	PriceCurrency string

	IsActive   bool
	IsArchived bool

	CreatedAt time.Time
	UpdatedAt time.Time

	Components []BundleComponent
}
