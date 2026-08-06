package service

import (
	"context"
	"fmt"
	"time"

	comtel "github.com/Krokozabra213/gargantua_common/pkg/telemetry"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/GargantuaLabs/order-product/internal/telemetry"
)

const productsListOp = "products.list"

func (s *Service) ListProducts(
	ctx context.Context,
	onlyActive bool,
) ([]domain.ProductWithComponents, error) {
	ctx, span := telemetry.ServiceTrace.Start(ctx, productsListOp)
	defer span.End()

	cacheKey := fmt.Sprintf("%s%v", domain.ProductsCacheKey, onlyActive)
	if cached, err := s.providers.Cache.GetProducts(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	v, err, _ := s.sf.Do(cacheKey, func() (any, error) {
		if cached, err := s.providers.Cache.GetProducts(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}

		products, err := s.providers.Product.ListWithComponents(ctx, onlyActive)
		if err != nil {
			comtel.HandleError(span, "failed list products", err)
			return nil, apperror.NewAppErr(
				apperror.CodeInternal,
				"repository.ListWithComponents",
				"failed list products",
				err,
				apperror.LevelError,
				nil,
			)
		}

		if len(products) > 0 {
			_ = s.providers.Cache.SetProducts(ctx, cacheKey, products, 20*time.Minute)
		}

		return products, nil
	})

	if err != nil {
		return nil, err
	}

	return v.([]domain.ProductWithComponents), nil
}
