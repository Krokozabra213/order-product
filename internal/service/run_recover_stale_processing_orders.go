package service

import (
	"context"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
)

func (s *Service) RunRecoverStaleProcessingOrders(
	ctx context.Context,
	timeout time.Duration,
	limit int,
) (int64, error) {
	count, err := s.providers.Order.RecoverStaleOrders(ctx, timeout, domain.OrderStatusProcessing, limit)
	if err != nil {
		s.metrics.RecoverProcessedOrdersFailedTotal.Add(ctx, 1)
		return 0, err
	}

	s.metrics.RecoverProcessedOrdersTotal.Add(ctx, count)
	return count, nil
}
