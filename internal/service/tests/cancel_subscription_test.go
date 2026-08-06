package tests

import (
	"context"
	"errors"
	"testing"

	apperror "github.com/GargantuaLabs/order-product/internal/app_error"
	"github.com/GargantuaLabs/order-product/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestService_CancelSubscription(t *testing.T) {

	type testCase struct {
		name        string
		setupMocks  func(deps *testDeps, userID, productID uuid.UUID)
		expectError bool
		expectCode  apperror.Code
	}

	tests := []testCase{
		{
			name: "success",
			setupMocks: func(deps *testDeps, userID, productID uuid.UUID) {
				subID := uuid.New()

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(&domain.Subscription{
						ID:     subID,
						Status: domain.SubscriptionStatusActive,
					}, nil)

				deps.sub.EXPECT().
					Cancel(gomock.Any(), subID).
					Return(nil)
			},
			expectError: false,
		},
		{
			name: "already canceled",
			setupMocks: func(deps *testDeps, userID, productID uuid.UUID) {
				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(nil, domain.ErrNotFound)
			},
			expectError: true,
			expectCode:  apperror.CodeNotFound,
		},
		{
			name: "get active error",
			setupMocks: func(deps *testDeps, userID, productID uuid.UUID) {
				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(nil, errors.New("db error"))
			},
			expectError: true,
			expectCode:  apperror.CodeInternal,
		},
		{
			name: "cancel error",
			setupMocks: func(deps *testDeps, userID, productID uuid.UUID) {
				subID := uuid.New()

				deps.uow.EXPECT().
					Do(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				deps.sub.EXPECT().
					GetActiveByUserAndProduct(gomock.Any(), userID, productID).
					Return(&domain.Subscription{
						ID:     subID,
						Status: domain.SubscriptionStatusActive,
					}, nil)

				deps.sub.EXPECT().
					Cancel(gomock.Any(), subID).
					Return(errors.New("db error"))
			},
			expectError: true,
			expectCode:  apperror.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			svc, deps := newTestService(t)
			defer deps.ctrl.Finish()

			ctx := context.Background()
			userID := uuid.New()
			productID := uuid.New()

			tt.setupMocks(deps, userID, productID)

			err := svc.CancelSubscription(ctx, userID, productID)

			if tt.expectError {
				require.Error(t, err)
				appErr := apperror.GetAppErr(err)

				require.NotNil(t, appErr)
				require.Equal(t, tt.expectCode, appErr.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
