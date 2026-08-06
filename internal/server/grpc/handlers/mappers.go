package handlers

import (
	"errors"

	orderproductv1 "github.com/Krokozabra213/gargantua_common/gen/go/proto/order-product/v1"
	"github.com/GargantuaLabs/order-product/internal/domain"
)

func mapToProtoOrderStatus(s domain.OrderStatus) orderproductv1.OrderStatus {
	switch s {
	case domain.OrderStatusPending:
		return orderproductv1.OrderStatus_ORDER_STATUS_PENDING
	case domain.OrderStatusCompleted:
		return orderproductv1.OrderStatus_ORDER_STATUS_COMPLETED
	case domain.OrderStatusFailed:
		return orderproductv1.OrderStatus_ORDER_STATUS_FAILED
	case domain.OrderStatusCancelled:
		return orderproductv1.OrderStatus_ORDER_STATUS_CANCELLED
	case domain.OrderStatusRefunded:
		return orderproductv1.OrderStatus_ORDER_STATUS_REFUNDED
	case domain.OrderStatusExpired:
		return orderproductv1.OrderStatus_ORDER_STATUS_EXPIRED
	default:
		return orderproductv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func mapToProtoProductType(t domain.ProductType) orderproductv1.ProductType {
	switch t {
	case domain.ProductTypeSubscription:
		return orderproductv1.ProductType_PRODUCT_TYPE_SUBSCRIPTION
	case domain.ProductTypeBundle:
		return orderproductv1.ProductType_PRODUCT_TYPE_BUNDLE
	default:
		return orderproductv1.ProductType_PRODUCT_TYPE_UNSPECIFIED
	}
}

func mapToProtoSubscriptionStatus(s domain.SubscriptionStatus) orderproductv1.SubscriptionStatus {
	switch s {
	case domain.SubscriptionStatusActive:
		return orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE
	case domain.SubscriptionStatusExpired:
		return orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_EXPIRED
	case domain.SubscriptionStatusCanceled:
		return orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED
	default:
		return orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED
	}
}

func mapToDomainSubscriptionStatus(s orderproductv1.SubscriptionStatus) (domain.SubscriptionStatus, error) {
	switch s {
	case orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE:

		return domain.SubscriptionStatusActive, nil
	case orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED:
		return domain.SubscriptionStatusCanceled, nil
	case orderproductv1.SubscriptionStatus_SUBSCRIPTION_STATUS_EXPIRED:
		return domain.SubscriptionStatusExpired, nil
	default:
		return domain.SubscriptionStatus(""), errors.New("unknown subscription status")
	}
}
