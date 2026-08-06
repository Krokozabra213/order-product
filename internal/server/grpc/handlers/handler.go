package handlers

import (
	"context"
	"time"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	"github.com/GargantuaLabs/order-product/internal/config"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	PurchaseProduct(ctx context.Context, userID, productID uuid.UUID) (*domain.Order, error)

	CancelSubscription(ctx context.Context, userID, productID uuid.UUID) error

	ListProducts(ctx context.Context, onlyActive bool) ([]domain.ProductWithComponents, error)

	ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.Order, error)

	ListUserSubscriptions(ctx context.Context, userID uuid.UUID, status domain.SubscriptionStatus) ([]domain.Subscription, error)

	GetOrder(ctx context.Context, orderID uuid.UUID) (*domain.Order, error)

	GetProduct(ctx context.Context, productID uuid.UUID) (*domain.ProductWithComponents, error)
}

type PaymentAPI struct {
	orderproductv1.OrderProductService_APIServer
	cfg     *config.Config
	service Service
}

func New(cfg *config.Config, svc Service) *PaymentAPI {
	return &PaymentAPI{
		cfg:     cfg,
		service: svc,
	}
}

func toProtoTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
