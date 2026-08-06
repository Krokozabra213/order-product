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

const subscriptionsListByUserOp = "subscriptions.list_by_user"

func (s *Service) ListUserSubscriptions(
	ctx context.Context,
	userID uuid.UUID,
	status domain.SubscriptionStatus,
) ([]domain.Subscription, error) {
	baseFields := apperror.Fields{
		apperror.F("user_id", userID.String()),
		apperror.F("status", string(status)),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, subscriptionsListByUserOp,
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("subscription.status", string(status)),
		),
	)
	defer span.End()

	subs, err := s.providers.Subscription.
		ListByUser(ctx, userID, status)
	if err != nil {
		comtel.HandleError(span, "failed list user subscriptions", err)
		return nil, apperror.NewAppErr(
			apperror.CodeInternal,
			"repository.ListByUser",
			"failed list user subscriptions",
			err,
			apperror.LevelError,
			baseFields,
		)
	}

	return subs, nil
}
