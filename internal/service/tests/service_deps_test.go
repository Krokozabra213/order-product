package tests

import (
	"log/slog"
	"testing"

	"github.com/GargantuaLabs/order-product/internal/config"
	"github.com/GargantuaLabs/order-product/internal/service"
	"github.com/GargantuaLabs/order-product/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

type testDeps struct {
	ctrl *gomock.Controller

	uow     *mocks.MockUnitOfWork
	cache   *mocks.MockCacheProvider
	product *mocks.MockProductProvider
	order   *mocks.MockOrderProvider
	sub     *mocks.MockSubscriptionProvider
	act     *mocks.MockSubscriptionActivationProvider
	balance *mocks.MockBalanceClient
}

func newTestService(t *testing.T) (*service.Service, *testDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)

	deps := &testDeps{
		ctrl:    ctrl,
		uow:     mocks.NewMockUnitOfWork(ctrl),
		cache:   mocks.NewMockCacheProvider(ctrl),
		product: mocks.NewMockProductProvider(ctrl),
		order:   mocks.NewMockOrderProvider(ctrl),
		sub:     mocks.NewMockSubscriptionProvider(ctrl),
		act:     mocks.NewMockSubscriptionActivationProvider(ctrl),
		balance: mocks.NewMockBalanceClient(ctrl),
	}

	svc := service.New(
		deps.uow,
		&config.Config{},
		slog.Default(),
		&service.Providers{
			Cache:                  deps.cache,
			Product:                deps.product,
			Order:                  deps.order,
			Subscription:           deps.sub,
			SubscriptionActivation: deps.act,
		},
		deps.balance,
	)

	return svc, deps
}
