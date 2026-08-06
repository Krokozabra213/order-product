package handlers

import (
	"context"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
)

func (h *PaymentAPI) ListProducts(
	ctx context.Context,
	req *orderproductv1.ListProductsRequest,
) (*orderproductv1.ListProductsResponse, error) {

	products, err := h.service.ListProducts(ctx, req.OnlyActive)
	if err != nil {
		return nil, err
	}

	protoProducts := make([]*orderproductv1.Product, 0, len(products))

	for _, p := range products {
		bundleItems := make([]*orderproductv1.BundleItem, 0, len(p.Components))
		for _, component := range p.Components {
			bundleItems = append(bundleItems, &orderproductv1.BundleItem{
				ProductId: component.ProductID.String(),
				Name:      component.Name,
			})
		}

		protoProducts = append(protoProducts, &orderproductv1.Product{
			Id:            p.ID.String(),
			Type:          mapToProtoProductType(p.Type),
			Name:          p.Name,
			Description:   p.Description,
			DurationDays:  int32(p.DurationDays),
			PriceAmount:   p.PriceAmount,
			PriceCurrency: p.PriceCurrency,
			BundleItems:   bundleItems,
		})
	}

	return &orderproductv1.ListProductsResponse{
		Products: protoProducts,
	}, nil
}
