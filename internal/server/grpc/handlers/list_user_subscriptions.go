package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const OpListUserSubscriptions = "handler.ListUserSubscriptions"

func (h *PaymentAPI) ListUserSubscriptions(
	ctx context.Context,
	req *orderproductv1.ListUserSubscriptionsRequest,
) (*orderproductv1.ListUserSubscriptionsResponse, error) {
	span := trace.SpanFromContext(ctx)
	baseFields := apperror.Fields{
		apperror.F("user_id", req.GetUserId()),
		apperror.F("status", req.GetStatus()),
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpListUserSubscriptions,
			"failed parse userID", err, apperror.LevelError, baseFields)
	}

	status, err := mapToDomainSubscriptionStatus(req.Status)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpListUserSubscriptions,
			"failed parse status", err, apperror.LevelError, baseFields)
	}

	subs, err := h.service.ListUserSubscriptions(ctx, userID, status)
	if err != nil {
		return nil, err
	}

	resp := &orderproductv1.ListUserSubscriptionsResponse{}

	for _, s := range subs {
		resp.Subscriptions = append(resp.Subscriptions, &orderproductv1.Subscription{
			Id:          s.ID.String(),
			ProductId:   s.ProductID.String(),
			ProductName: s.ProductName,
			Status:      mapToProtoSubscriptionStatus(s.Status),
			ExpiresAt:   timestamppb.New(s.ExpiresAt),
		})
	}

	return resp, nil
}
