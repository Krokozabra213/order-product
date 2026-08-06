package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/GargantuaLabs/order-product/internal/config"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/GargantuaLabs/order-product/internal/telemetry"
	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
)

const componentName = "internal/service"

//go:generate mockgen -source=Service.go -destination=mocks/mock_service.go -package=mocks -typed

type CacheProvider interface {
	GetProducts(ctx context.Context, key string) ([]domain.ProductWithComponents, error)
	SetProducts(ctx context.Context, key string, value []domain.ProductWithComponents, ttl time.Duration) error
	GetProduct(ctx context.Context, key string) (*domain.ProductWithComponents, error)
	SetProduct(ctx context.Context, key string, value *domain.ProductWithComponents, ttl time.Duration) error
}

type ProductProvider interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetBundleComponents(ctx context.Context, bundleID uuid.UUID) ([]domain.Product, error)
	ListWithComponents(ctx context.Context, onlyActive bool) ([]domain.ProductWithComponents, error)
	GetWithComponents(ctx context.Context, productID uuid.UUID) (*domain.ProductWithComponents, error)
}

type OrderProvider interface {
	Create(ctx context.Context, userID uuid.UUID, product *domain.Product) (*domain.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.Order, error)
	TakePendingBatch(ctx context.Context, limit int) ([]domain.Order, error)
	RecoverStaleOrders(ctx context.Context, timeout time.Duration, status domain.OrderStatus, limit int) (int64, error)
}

type SubscriptionProvider interface {
	GetActiveByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*domain.Subscription, error)
	Create(ctx context.Context, userID, productID uuid.UUID, productName string, expAt time.Time) (*domain.Subscription, error)
	UpdateExpiresAt(ctx context.Context, subId uuid.UUID, expAt time.Time) error
	ListByUser(ctx context.Context, userID uuid.UUID, status domain.SubscriptionStatus) ([]domain.Subscription, error)
	Cancel(ctx context.Context, subscriptionID uuid.UUID) error
}

type SubscriptionActivationProvider interface {
	Create(ctx context.Context, subID, orderID uuid.UUID, duration int, startedAt time.Time, expAt time.Time) error
}

type BalanceClient interface {
	Charge(ctx context.Context, operationID string, userID, orderID uuid.UUID, amount int64) (*domain.BalanceStatus, error)
}

type Providers struct {
	Cache                  CacheProvider
	Product                ProductProvider
	Order                  OrderProvider
	Subscription           SubscriptionProvider
	SubscriptionActivation SubscriptionActivationProvider
}

func NewProviders(
	cache CacheProvider,
	product ProductProvider,
	order OrderProvider,
	subscription SubscriptionProvider,
	subscriptionActivation SubscriptionActivationProvider,
) *Providers {
	return &Providers{
		Cache:                  cache,
		Product:                product,
		Order:                  order,
		Subscription:           subscription,
		SubscriptionActivation: subscriptionActivation,
	}
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	uow       UnitOfWork
	cfg       *config.Config
	log       *slog.Logger
	providers *Providers
	balance   BalanceClient
	sf        singleflight.Group
	metrics   *telemetry.OrderProductMetrics
}

func New(
	uow UnitOfWork,
	cfg *config.Config,
	log *slog.Logger,
	providers *Providers,
	balance BalanceClient,
) *Service {
	return &Service{
		uow:       uow,
		cfg:       cfg,
		log:       log.With(slog.String("component", componentName)),
		providers: providers,
		balance:   balance,
		metrics:   telemetry.NewOrderProductMetrics(telemetry.ServiceMeter),
	}
}
