package domain

import (
	"time"

	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusCompleted  OrderStatus = "COMPLETED"
	OrderStatusFailed     OrderStatus = "FAILED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
	OrderStatusRefunded   OrderStatus = "REFUNDED"
	OrderStatusExpired    OrderStatus = "EXPIRED"
)

func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusPending,
		OrderStatusProcessing,
		OrderStatusCompleted,
		OrderStatusFailed,
		OrderStatusCancelled,
		OrderStatusRefunded,
		OrderStatusExpired:
		return true
	default:
		return false
	}
}

type Order struct {
	ID     uuid.UUID
	UserID uuid.UUID

	ProductID   uuid.UUID
	ProductName string
	ProductType ProductType

	DurationDays int

	Amount   types.PositiveInt[int64]
	Currency CurrencyType

	Status OrderStatus

	BalanceOperationID *uuid.UUID

	PaidAt      *time.Time
	CancelledAt *time.Time
	ExpiredAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
