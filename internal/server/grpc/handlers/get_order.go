package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

const OpGetOrder = "handler.GetOrder"

func (h *PaymentAPI) GetOrder(
	ctx context.Context,
	req *orderproductv1.GetOrderRequest,
) (*orderproductv1.GetOrderResponse, error) {
	span := trace.SpanFromContext(ctx)
	baseFields := apperror.Fields{
		apperror.F("order_id", req.GetOrderId()),
	}

	orderID, err := uuid.Parse(req.OrderId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpGetOrder,
			"failed parse orderID", err, apperror.LevelError, baseFields)
	}

	order, err := h.service.GetOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &orderproductv1.GetOrderResponse{
		OrderId:   order.ID.String(),
		UserId:    order.UserID.String(),
		ProductId: order.ProductID.String(),
		Status:    mapToProtoOrderStatus(order.Status),
		Amount:    order.Amount.Value(),
		Currency:  string(order.Currency),
	}, nil
}
