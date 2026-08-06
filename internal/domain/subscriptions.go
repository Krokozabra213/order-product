package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "ACTIVE"
	SubscriptionStatusExpired  SubscriptionStatus = "EXPIRED"
	SubscriptionStatusCanceled SubscriptionStatus = "CANCELED"
)

func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionStatusActive,
		SubscriptionStatusExpired,
		SubscriptionStatusCanceled:
		return true
	default:
		return false
	}
}

type Subscription struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ProductID   uuid.UUID
	ProductName string

	Status SubscriptionStatus

	ExpiresAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
