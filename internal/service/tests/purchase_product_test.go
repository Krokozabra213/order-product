package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/Krokozabra213/common/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_PurchaseProduct(t *testing.T) {

	type testCase struct {
		name  string
		setup func(deps *testDeps, userID, productID uuid.UUID)

		expectErr        bool
		expectAppErrCode apperror.Code
		expectOrderNil   bool
	}

	tests := []testCase{

		{
			name: "product not found",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {
				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(nil, domain.ErrNotFound)
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeNotFound,
			expectOrderNil:   true,
		},

		{
			name: "product archived",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {
				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(&domain.Product{
						ID:         productID,
						IsActive:   false,
						IsArchived: true,
					}, nil)
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeInvalidArgument,
			expectOrderNil:   true,
		},

		{
			name: "double click - order already exists",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(nil, domain.ErrAlreadyExists)
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeAlreadyExists,
			expectOrderNil:   true,
		},

		{
			name: "balance error",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					DurationDays: 30,
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)
				statusFailed := domain.BalanceStatusFailed
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusFailed, errors.New("balance down"))
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeAborted,
			expectOrderNil:   true,
		},

		{
			name: "balance no money",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					DurationDays: 30,
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:     uuid.New(),
					UserID: userID,
					Amount: amount,

					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusNoMoney := domain.BalanceStatusNoMoney
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusNoMoney, nil)

				deps.order.EXPECT().
					UpdateStatus(gomock.Any(), order.ID, domain.OrderStatusFailed).
					Return(nil)
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeInsufficientFunds,
			expectOrderNil:   false,
		},

		{
			name: "happy path subscription",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					Type:         domain.ProductTypeSubscription,
					Name:         "Pro",
					DurationDays: 30,
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusSuccess := domain.BalanceStatusSuccess
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusSuccess, nil)

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.order.EXPECT().
					UpdateStatus(gomock.Any(), order.ID, domain.OrderStatusCompleted).
					Return(nil)

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(nil, domain.ErrNotFound)

				deps.sub.EXPECT().
					Create(gomock.Any(), userID, productID, product.Name, gomock.Any()).
					Return(&domain.Subscription{ID: uuid.New()}, nil)

				deps.act.EXPECT().
					Create(gomock.Any(), gomock.Any(), order.ID, order.DurationDays, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr:      false,
			expectOrderNil: false,
		},

		{
			name: "happy path bundle",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					Type:         domain.ProductTypeBundle,
					Name:         "Bundle",
					DurationDays: 30,
				}

				child1 := &domain.Product{
					ID:   uuid.New(),
					Name: "Sub1",
				}
				child2 := &domain.Product{
					ID:   uuid.New(),
					Name: "Sub2",
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusSuccess := domain.BalanceStatusSuccess
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusSuccess, nil)

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.order.EXPECT().
					UpdateStatus(gomock.Any(), order.ID, domain.OrderStatusCompleted).
					Return(nil)

				deps.product.EXPECT().
					GetBundleComponents(gomock.Any(), productID).
					Return([]domain.Product{*child1, *child2}, nil)

				for _, child := range []*domain.Product{child1, child2} {
					deps.sub.EXPECT().
						GetActiveByUserAndProduct(gomock.Any(), userID, child.ID).
						Return(nil, domain.ErrNotFound)

					deps.sub.EXPECT().
						Create(gomock.Any(), userID, child.ID, child.Name, gomock.Any()).
						Return(&domain.Subscription{ID: uuid.New()}, nil)

					deps.act.EXPECT().
						Create(gomock.Any(), gomock.Any(), order.ID, order.DurationDays, gomock.Any(), gomock.Any()).
						Return(nil)
				}
			},
			expectErr: false,
		},

		{
			name: "activation error rollback",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					Type:         domain.ProductTypeSubscription,
					Name:         "Pro",
					DurationDays: 30,
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusSuccess := domain.BalanceStatusSuccess
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusSuccess, nil)

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.order.EXPECT().
					UpdateStatus(gomock.Any(), order.ID, domain.OrderStatusCompleted).
					Return(nil)

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(nil, domain.ErrNotFound)

				deps.sub.EXPECT().
					Create(gomock.Any(), userID, productID, product.Name, gomock.Any()).
					Return(&domain.Subscription{ID: uuid.New()}, nil)

				deps.act.EXPECT().
					Create(gomock.Any(), gomock.Any(), order.ID, order.DurationDays, gomock.Any(), gomock.Any()).
					Return(errors.New("activation failed"))
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeAborted,
			expectOrderNil:   true,
		},

		{
			name: "subscription renewal",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					Type:         domain.ProductTypeSubscription,
					Name:         "Pro",
					DurationDays: 30,
				}

				existingSub := &domain.Subscription{
					ID:        uuid.New(),
					ExpiresAt: time.Now().Add(10 * 24 * time.Hour),
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusSuccess := domain.BalanceStatusSuccess
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusSuccess, nil)

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.order.EXPECT().
					UpdateStatus(gomock.Any(), order.ID, domain.OrderStatusCompleted).
					Return(nil)

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(existingSub, nil)

				deps.sub.EXPECT().
					UpdateExpiresAt(gomock.Any(), existingSub.ID, gomock.Any()).
					Return(nil)

				deps.act.EXPECT().
					Create(gomock.Any(), existingSub.ID, order.ID, order.DurationDays, gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectErr: false,
		},

		{
			name: "uow error",
			setup: func(deps *testDeps, userID, productID uuid.UUID) {

				product := &domain.Product{
					ID:           productID,
					IsActive:     true,
					IsArchived:   false,
					Type:         domain.ProductTypeSubscription,
					Name:         "Pro",
					DurationDays: 30,
				}

				amount, _ := types.NewPositiveInt(int64(1000))

				order := &domain.Order{
					ID:           uuid.New(),
					UserID:       userID,
					Amount:       amount,
					DurationDays: 30,
				}

				deps.product.EXPECT().
					GetByID(gomock.Any(), productID).
					Return(product, nil)

				deps.order.EXPECT().
					Create(gomock.Any(), userID, product).
					Return(order, nil)

				statusSuccess := domain.BalanceStatusSuccess
				deps.balance.EXPECT().
					Charge(gomock.Any(), gomock.Any(), userID, order.ID, order.Amount.Value()).
					Return(&statusSuccess, nil)

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					Return(errors.New("tx fail"))
			},
			expectErr:        true,
			expectAppErrCode: apperror.CodeAborted,
			expectOrderNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc, deps := newTestService(t)
			defer deps.ctrl.Finish()

			ctx := context.Background()
			userID := uuid.New()
			productID := uuid.New()

			tt.setup(deps, userID, productID)

			order, err := svc.PurchaseProduct(ctx, userID, productID)

			if tt.expectErr {
				require.Error(t, err)
				require.Error(t, err)
				appErr := apperror.GetAppErr(err)

				require.NotNil(t, appErr)
				if tt.expectAppErrCode != 0 {
					require.Equal(t, tt.expectAppErrCode, appErr.Code())
				}
			} else {
				require.NoError(t, err)
			}

			if tt.expectOrderNil {
				require.Nil(t, order)
			}
		})
	}
}
