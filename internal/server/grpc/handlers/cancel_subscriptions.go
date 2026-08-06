package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

const OpCancelSubscription = "handler.CancelSubscription"

func (h *PaymentAPI) CancelSubscription(
	ctx context.Context,
	req *orderproductv1.CancelSubscriptionRequest,
) (*orderproductv1.CancelSubscriptionResponse, error) {
	span := trace.SpanFromContext(ctx)
	baseFields := apperror.Fields{
		apperror.F("user_id", req.GetUserId()),
		apperror.F("product_id", req.GetProductId()),
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpCancelSubscription,
			"failed parse userID", err, apperror.LevelError, baseFields)
	}

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpCancelSubscription,
			"failed parse productID", err, apperror.LevelError, baseFields)
	}

	err = h.service.CancelSubscription(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	return &orderproductv1.CancelSubscriptionResponse{
		Success: true,
	}, nil
}
