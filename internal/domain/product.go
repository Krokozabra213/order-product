package domain

import (
	"time"

	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
)

type ProductType string

const (
	ProductTypeSubscription ProductType = "SUBSCRIPTION"
	ProductTypeBundle       ProductType = "BUNDLE"
)

func (t ProductType) IsValid() bool {
	switch t {
	case ProductTypeSubscription,
		ProductTypeBundle:
		return true
	default:
		return false
	}
}

type CurrencyType string

const (
	CurrencyTypeRub CurrencyType = "RUB"
)

type Product struct {
	ID           uuid.UUID
	Type         ProductType
	Name         string
	Description  string
	DurationDays int

	PriceAmount   types.PositiveInt[int64]
	PriceCurrency CurrencyType

	IsActive   bool
	IsArchived bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProductComponent struct {
	ParentProductID uuid.UUID
	ChildProductID  uuid.UUID
}
