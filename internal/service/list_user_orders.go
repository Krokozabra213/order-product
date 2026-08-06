package service

import (
	"context"

	comtel "github.com/Krokozabra213/gargantua_common/pkg/telemetry"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/GargantuaLabs/order-product/internal/telemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const ordersListOp = "orders.list"

func (s *Service) ListUserOrders(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int32,
) ([]domain.Order, error) {
	baseFields := apperror.Fields{
		apperror.F("user_id", userID.String()),
		apperror.F("limit", limit),
		apperror.F("offset", offset),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, ordersListOp,
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.Int("limit", int(limit)),
			attribute.Int("offset", int(offset)),
		),
	)
	defer span.End()

	orders, err := s.providers.Order.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		comtel.HandleError(span, "failed list user orders", err)
		return nil, apperror.NewAppErr(
			apperror.CodeInternal,
			"repository.ListByUser",
			"failed list user orders",
			err,
			apperror.LevelError,
			baseFields,
		)
	}
	return orders, nil
}
