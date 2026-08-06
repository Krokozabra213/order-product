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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const purchaseProductOp = "product.purchase"

func (s *Service) PurchaseProduct(
	ctx context.Context,
	userID, productID uuid.UUID,
) (*domain.Order, error) {

	baseFields := apperror.Fields{
		apperror.F("user_id", userID.String()),
		apperror.F("product_id", productID.String()),
	}

	ctx, span := telemetry.ServiceTrace.Start(ctx, purchaseProductOp,
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("product.id", productID.String()),
		),
	)
	defer span.End()

	s.metrics.PurchaseProductInProcess.Add(ctx, 1)
	defer s.metrics.PurchaseProductInProcess.Add(ctx, -1)

	product, err := s.providers.Product.GetByID(ctx, productID)
	if err != nil {
		comtel.HandleError(span, "failed get product", err)
		s.metrics.PurchaseProductFailed.Add(ctx, 1)

		code := apperror.CodeInternal
		if errors.Is(err, domain.ErrNotFound) {
			code = apperror.CodeNotFound
		}

		return nil, apperror.NewAppErr(
			code,
			"repository.Product.GetByID",
			"failed get product",
			err,
			apperror.LevelError,
			baseFields,
		)
	}

	if !product.IsActive || product.IsArchived {
		s.metrics.PurchaseProductFailed.Add(ctx, 1)
		return nil, apperror.NewAppErr(
			apperror.CodeInvalidArgument,
			"product.isActive",
			"product is archived",
			err,
			apperror.LevelError,
			baseFields,
		)
	}

	order, err := s.providers.Order.Create(ctx, userID, product)
	if err != nil {
		comtel.HandleError(span, "failed create order", err)
		s.metrics.PurchaseProductFailed.Add(ctx, 1)

		code := apperror.CodeInternal
		if errors.Is(err, domain.ErrAlreadyExists) {
			code = apperror.CodeAlreadyExists
		}

		return nil, apperror.NewAppErr(
			code,
			"repository.Order.Create",
			"failed create order",
			err,
			apperror.LevelError,
			baseFields,
		)
	}

	err = s.ProcessPendingOrder(ctx, order, product, baseFields)
	if err != nil {
		comtel.HandleError(span, "failed process pending order", err)
		s.metrics.PurchaseProductFailed.Add(ctx, 1)
		return nil, err
	}

	s.metrics.PurchaseProductCompleted.Add(ctx, 1)
	return order, nil
}

func (s *Service) ProcessPendingOrder(
	ctx context.Context,
	order *domain.Order,
	product *domain.Product,
	fields apperror.Fields,
) error {

	operationID := "order-product:withdraw:" + order.ID.String()

	status, err := s.balance.Charge(ctx, operationID, order.UserID, order.ID, order.Amount.Value())
	if err != nil {
		return apperror.NewAppErr(
			apperror.CodeAborted,
			"balance.Charge",
			"failed withdraw balance",
			err,
			apperror.LevelError,
			fields,
		)
	}

	if *status == domain.BalanceStatusFailed || *status == domain.BalanceStatusNoMoney {

		_ = s.providers.Order.UpdateStatus(ctx, order.ID, domain.OrderStatusFailed)

		code := apperror.CodeInternal
		if *status == domain.BalanceStatusNoMoney {
			code = apperror.CodeInsufficientFunds
		}

		return apperror.NewAppErr(
			code,
			"balance.status",
			"failed withdraw balance",
			nil,
			apperror.LevelError,
			nil,
		)
	}

	err = s.uow.Do(ctx, func(ctx context.Context) error {

		if err := s.providers.Order.UpdateStatus(
			ctx,
			order.ID,
			domain.OrderStatusCompleted,
		); err != nil {
			return fmt.Errorf("failed update order status: %w", err)
		}

		switch product.Type {

		case domain.ProductTypeSubscription:
			return s.activateSubscription(ctx, order.UserID, product, order)

		case domain.ProductTypeBundle:

			components, err := s.providers.Product.
				GetBundleComponents(ctx, product.ID)
			if err != nil {
				return fmt.Errorf("failed get bundle product components: %w", err)
			}

			for _, childProduct := range components {
				if err := s.activateSubscription(ctx, order.UserID, &childProduct, order); err != nil {
					return fmt.Errorf("failed activate subscription: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return apperror.NewAppErr(
			apperror.CodeAborted,
			"s.uow.Do",
			"operation partially completed, remaining steps will be finished by background worker within a few minutes",
			err,
			apperror.LevelError,
			fields,
		)
	}
	return nil

}

func (s *Service) activateSubscription(
	ctx context.Context,
	userID uuid.UUID, product *domain.Product,
	order *domain.Order,
) error {
	now := time.Now()

	sub, err := s.providers.Subscription.
		GetActiveByUserAndProduct(ctx, userID, product.ID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return errors.New("failed get active subscription")
	}

	var (
		startedAt time.Time
		expiresAt time.Time
	)

	if sub == nil {
		startedAt = now
		expiresAt = now.AddDate(0, 0, order.DurationDays)

		sub, err = s.providers.Subscription.Create(ctx, userID, product.ID, product.Name, expiresAt)
		if err != nil {
			return fmt.Errorf("failed create subscription: %w", err)
		}

	} else {
		if sub.ExpiresAt.After(now) {
			startedAt = sub.ExpiresAt
		} else {
			startedAt = now
		}

		expiresAt = startedAt.AddDate(0, 0, order.DurationDays)

		err = s.providers.Subscription.UpdateExpiresAt(ctx, sub.ID, expiresAt)
		if err != nil {
			return fmt.Errorf("failed update subscription: %w", err)
		}
	}

	err = s.providers.SubscriptionActivation.Create(ctx, sub.ID, order.ID, order.DurationDays, startedAt, expiresAt)
	if err != nil {
		return errors.New("failed create subscription activation")
	}
	return nil
}
