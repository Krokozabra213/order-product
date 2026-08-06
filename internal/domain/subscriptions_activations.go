package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionActivation struct {
	ID             uuid.UUID
	SubscriptionID uuid.UUID
	OrderID        uuid.UUID

	DurationDays int

	StartedAt time.Time
	ExpiresAt time.Time

	CreatedAt time.Time
}
