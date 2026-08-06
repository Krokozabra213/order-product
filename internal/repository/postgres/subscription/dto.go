package subRepo

import (
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
)

type SubscriptionRow struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	ProductID   uuid.UUID `db:"product_id"`
	ProductName string    `db:"product_name"`
	ExpiresAt   time.Time `db:"expires_at"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (r *SubscriptionRow) ToDomain() (*domain.Subscription, error) {
	status := domain.SubscriptionStatus(r.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid subscription status: %s", r.Status)
	}

	return &domain.Subscription{
		ID:          r.ID,
		UserID:      r.UserID,
		ProductID:   r.ProductID,
		ProductName: r.ProductName,
		ExpiresAt:   r.ExpiresAt,
		Status:      status,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}, nil
}
