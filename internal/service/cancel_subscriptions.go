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
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const subscriptionCancelOp = "subscription.cancel"

func (s *Service) CancelSubscription(
	ctx context.Context,
	userID, productID uuid.UUID,
) error {
	baseFields := apperror.Fields{
		apperror.F("user_id", userID.String()),
		apperror.F("product_id", productID.String()),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, subscriptionCancelOp,
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("product.id", productID.String()),
		),
	)
	defer span.End()

	s.metrics.CancelSubscriptionsInProcess.Add(ctx, 1)
	defer s.metrics.CancelSubscriptionsInProcess.Add(ctx, -1)

	return s.uow.Do(ctx, func(ctx context.Context) error {

		sub, err := s.providers.Subscription.
			GetActiveByUserAndProduct(ctx, userID, productID)
		if err != nil {
			comtel.HandleError(span, "failed get active subscriptions", err)
			s.metrics.CancelSubscriptionsFailed.Add(ctx, 1, metric.WithAttributes(
				attribute.String("user.id", userID.String()),
				attribute.String("product.id", productID.String()),
				attribute.String("error.type", comtel.DBAttr),
			))

			code := apperror.CodeInternal
			if errors.Is(err, domain.ErrNotFound) {
				code = apperror.CodeNotFound
			}

			return apperror.NewAppErr(
				code,
				"repository.Subscription.GetActiveByUserAndProduct",
				"failed get active subscription",
				err,
				apperror.LevelError,
				baseFields,
			)
		}

		err = s.providers.Subscription.Cancel(ctx, sub.ID)
		if err != nil {
			comtel.HandleError(span, "failed cancel subscription", err)
			s.metrics.CancelSubscriptionsFailed.Add(ctx, 1, metric.WithAttributes(
				attribute.String("user.id", userID.String()),
				attribute.String("product.id", productID.String()),
				attribute.String("error.type", comtel.DBAttr),
			))
			return apperror.NewAppErr(apperror.CodeInternal, "repository.Cancel",
				"failed cancel subscription", err, apperror.LevelError, baseFields)
		}
		s.metrics.CancelSubscriptionsCompleted.Add(ctx, 1, metric.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("product.id", productID.String()),
		))
		return nil
	})
}
