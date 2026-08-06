package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/GargantuaLabs/order-product/internal/domain"
)

func (s *Service) RunPendingOrdersWorker(ctx context.Context, limit int) error {
	start := time.Now()
	s.metrics.PendingOrdersInProcess.Add(ctx, 1)
	defer s.metrics.PendingOrdersInProcess.Add(ctx, -1)
	defer s.metrics.PendingOrdersProcessingDuration.Record(ctx, time.Since(start).Seconds())

	var orders []domain.Order

	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		orders, err = s.providers.Order.TakePendingBatch(ctx, limit)
		return err
	})

	if err != nil {
		s.metrics.PendingOrdersFailedTotal.Add(ctx, 1)
		return fmt.Errorf("list pending orders: %w", err)
	}

	for _, order := range orders {

		product, err := s.providers.Product.GetByID(ctx, order.ProductID)
		if err != nil {
			s.metrics.PendingOrdersFailedTotal.Add(ctx, 1)
			s.log.ErrorContext(ctx, "failed to get product, skipping order", slog.Any("error", err))
			continue
		}

		err = s.ProcessPendingOrder(ctx, &order, product, nil)
		if err != nil {
			s.metrics.PendingOrdersFailedTotal.Add(ctx, 1)
			s.log.Error("failed process order",
				"order_id", order.ID,
				"error", err,
			)
			continue
		}
		s.metrics.PendingOrdersProcessedTotal.Add(ctx, 1)
	}

	return nil
}
