package orderRepo

import (
	"fmt"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
)

type OrderRow struct {
	ID                 uuid.UUID  `db:"id"`
	UserID             uuid.UUID  `db:"user_id"`
	ProductID          uuid.UUID  `db:"product_id"`
	ProductName        string     `db:"product_name"`
	ProductType        string     `db:"product_type"`
	DurationDays       int        `db:"duration_days"`
	Amount             int64      `db:"amount"`
	Currency           string     `db:"currency"`
	Status             string     `db:"status"`
	BalanceOperationID *uuid.UUID `db:"balance_operation_id"`
	PaidAt             *time.Time `db:"paid_at"`
	CanceledAt         *time.Time `db:"canceled_at"`
	ExpiredAt          *time.Time `db:"expired_at"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

func (r *OrderRow) ToDomain() (*domain.Order, error) {
	productType := domain.ProductType(r.ProductType)
	if !productType.IsValid() {
		return nil, fmt.Errorf("invalid product type %q", r.ProductType)
	}

	status := domain.OrderStatus(r.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid status %q", r.Status)
	}

	amount, err := types.NewPositiveInt(r.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount %d: %w", r.Amount, err)
	}

	return &domain.Order{
		ID:                 r.ID,
		UserID:             r.UserID,
		ProductID:          r.ProductID,
		ProductName:        r.ProductName,
		ProductType:        productType,
		DurationDays:       r.DurationDays,
		Amount:             amount,
		Currency:           domain.CurrencyType(r.Currency),
		Status:             status,
		BalanceOperationID: r.BalanceOperationID,
		PaidAt:             r.PaidAt,
		CancelledAt:        r.CanceledAt,
		ExpiredAt:          r.ExpiredAt,
		CreatedAt:          r.CreatedAt,
		UpdatedAt:          r.UpdatedAt,
	}, nil
}
