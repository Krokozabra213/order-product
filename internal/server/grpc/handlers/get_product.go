package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

func (h *PaymentAPI) GetProduct(
	ctx context.Context,
	req *orderproductv1.GetProductRequest,
) (*orderproductv1.GetProductResponse, error) {
	span := trace.SpanFromContext(ctx)
	baseFields := apperror.Fields{
		apperror.F("product_id", req.GetProductId()),
	}

	productID, err := uuid.Parse(req.ProductId)
	if err != nil {
		span.RecordError(err)
		return nil, apperror.NewAppErr(apperror.CodeValidation, OpGetOrder,
			"failed parse productID", err, apperror.LevelError, baseFields)
	}

	product, err := h.service.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	bundleItems := make([]*orderproductv1.BundleItem, 0, len(product.Components))
	for _, component := range product.Components {
		bundleItems = append(bundleItems, &orderproductv1.BundleItem{
			ProductId: component.ProductID.String(),
			Name:      component.Name,
		})
	}

	return &orderproductv1.GetProductResponse{
		Product: &orderproductv1.Product{
			Id:            product.ID.String(),
			Type:          mapToProtoProductType(product.Type),
			Name:          product.Name,
			Description:   product.Description,
			DurationDays:  int32(product.DurationDays),
			PriceAmount:   product.PriceAmount,
			PriceCurrency: product.PriceCurrency,
			BundleItems:   bundleItems,
		},
	}, nil
}
