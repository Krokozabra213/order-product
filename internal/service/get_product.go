package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	comtel "github.com/Krokozabra213/gargantua_common/pkg/telemetry"
	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/GargantuaLabs/order-product/internal/telemetry"
	"github.com/google/uuid"
)

func (s *Service) GetProduct(
	ctx context.Context,
	productID uuid.UUID,
) (*domain.ProductWithComponents, error) {

	baseFields := apperror.Fields{
		apperror.F("product_id", productID),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, "products.GetProduct")
	defer span.End()

	cacheKey := fmt.Sprintf("%s:%s", domain.ProductCacheKey, productID.String())

	if cached, err := s.providers.Cache.GetProduct(ctx, cacheKey); err == nil && cached != nil {
		return cached, nil
	}

	v, err, _ := s.sf.Do(cacheKey, func() (any, error) {

		if cached, err := s.providers.Cache.GetProduct(ctx, cacheKey); err == nil && cached != nil {
			return cached, nil
		}

		product, err := s.providers.Product.GetWithComponents(ctx, productID)
		if err != nil {

			code := apperror.CodeInternal
			comtel.HandleError(span, "failed get product", err)
			if errors.Is(err, domain.ErrNotFound) {
				code = apperror.CodeNotFound
				return nil, err
			}

			return nil, apperror.NewAppErr(
				code,
				"repository.GetWithComponents",
				"failed get product",
				err,
				apperror.LevelError,
				baseFields,
			)
		}

		_ = s.providers.Cache.SetProduct(ctx, cacheKey, product, 20*time.Minute)

		return product, nil
	})

	if err != nil {
		return nil, err
	}

	return v.(*domain.ProductWithComponents), nil
}
