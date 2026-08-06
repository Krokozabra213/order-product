package service

import (
	"context"
	"errors"

	comtel "github.com/Krokozabra213/gargantua_common/pkg/telemetry"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/GargantuaLabs/order-product/internal/telemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const orderGetOp = "order.get"

func (s *Service) GetOrder(
	ctx context.Context,
	orderID uuid.UUID,
) (*domain.Order, error) {
	baseFields := apperror.Fields{
		apperror.F("order_id", orderID.String()),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, orderGetOp,
		trace.WithAttributes(
			attribute.String("order.id", orderID.String()),
		),
	)
	defer span.End()

	order, err := s.providers.Order.GetByID(ctx, orderID)
	if err != nil {
		comtel.HandleError(span, "failed get order by id", err)

		code := apperror.CodeInternal
		if errors.Is(err, domain.ErrNotFound) {
			code = apperror.CodeNotFound
		}

		return nil, apperror.NewAppErr(
			code,
			"repository.GetByID",
			"failed get order by id",
			err,
			apperror.LevelError,
			baseFields,
		)
	}

	return order, nil
}
