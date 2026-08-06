package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxLimit = 100
)

const OpListUserOrders = "handler.ListUserOrders"

func (h *PaymentAPI) ListUserOrders(
	ctx context.Context,
	req *orderproductv1.ListUserOrdersRequest,
) (*orderproductv1.ListUserOrdersResponse, error) {
	span := trace.SpanFromContext(ctx)
	baseFields := apperror.Fields{
		apperror.F("user_id", req.GetUserId()),
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpListUserOrders,
			"failed parse userID", err, apperror.LevelError, baseFields)
	}

	limit := req.GetLimit()
	if limit <= 0 || limit > maxLimit {
		limit = 20
	}

	orders, err := h.service.ListUserOrders(
		ctx,
		userID,
		limit,
		req.Offset,
	)
	if err != nil {
		return nil, err
	}

	resp := &orderproductv1.ListUserOrdersResponse{}

	for _, o := range orders {

		resp.Orders = append(resp.Orders, &orderproductv1.Order{
			Id:           o.ID.String(),
			ProductId:    o.ProductID.String(),
			ProductName:  o.ProductName,
			ProductType:  mapToProtoProductType(o.ProductType),
			DurationDays: int32(o.DurationDays),
			Amount:       o.Amount.Value(),
			Currency:     string(o.Currency),
			Status:       mapToProtoOrderStatus(o.Status),
			CreatedAt:    timestamppb.New(o.CreatedAt),
			PaidAt:       toProtoTime(o.PaidAt),
		})
	}

	return resp, nil
}
