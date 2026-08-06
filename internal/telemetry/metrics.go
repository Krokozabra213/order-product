package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	ServiceMeter = otel.Meter("order-product/internal/service")
	ServiceTrace = otel.Tracer("order-product/internal/service")
)

type OrderProductMetrics struct {
	CancelSubscriptionsFailed    metric.Int64Counter
	CancelSubscriptionsCompleted metric.Int64Counter
	CancelSubscriptionsInProcess metric.Int64UpDownCounter

	PurchaseProductFailed    metric.Int64Counter
	PurchaseProductCompleted metric.Int64Counter
	PurchaseProductInProcess metric.Int64UpDownCounter

	PendingOrdersProcessedTotal     metric.Int64Counter
	PendingOrdersFailedTotal        metric.Int64Counter
	PendingOrdersInProcess          metric.Int64UpDownCounter
	PendingOrdersProcessingDuration metric.Float64Histogram

	RecoverProcessedOrdersTotal       metric.Int64Counter
	RecoverProcessedOrdersFailedTotal metric.Int64Counter
}

func NewOrderProductMetrics(meter metric.Meter) *OrderProductMetrics {
	m := &OrderProductMetrics{}
	var err error

	m.CancelSubscriptionsFailed, err = meter.Int64Counter(
		"order_product.subscriptions.cancel.failed",
		metric.WithDescription("Number of failed subscription cancellations"),
		metric.WithUnit("{call}"),
	)
	must(err)

	m.CancelSubscriptionsCompleted, err = meter.Int64Counter(
		"order_product.subscriptions.cancel.completed",
		metric.WithDescription("Number of completed subscription cancellations"),
		metric.WithUnit("{call}"),
	)
	must(err)

	m.CancelSubscriptionsInProcess, err = meter.Int64UpDownCounter(
		"order_product.subscriptions.cancel.in_process",
		metric.WithDescription("Number of subscription cancellations currently in process"),
		metric.WithUnit("{call}"),
	)
	must(err)

	m.PurchaseProductFailed, err = meter.Int64Counter(
		"order_product.purchase.failed",
		metric.WithDescription("Number of failed product purchases"),
		metric.WithUnit("{call}"),
	)
	must(err)

	m.PurchaseProductCompleted, err = meter.Int64Counter(
		"order_product.purchase.completed",
		metric.WithDescription("Number of completed product purchases"),
		metric.WithUnit("{call}"),
	)
	must(err)

	m.PurchaseProductInProcess, err = meter.Int64UpDownCounter(
		"order_product.purchase.in_process",
		metric.WithDescription("Number of product purchases currently in process"),
		metric.WithUnit("{call}"),
	)

	m.PendingOrdersProcessedTotal, err = meter.Int64Counter(
		"order_product.pending_orders.processed_total",
		metric.WithDescription("Number of successfully processed pending orders"),
		metric.WithUnit("{order}"),
	)
	must(err)

	m.PendingOrdersFailedTotal, err = meter.Int64Counter(
		"order_product.pending_orders.failed_total",
		metric.WithDescription("Number of failed pending orders processing attempts"),
		metric.WithUnit("{order}"),
	)
	must(err)

	m.PendingOrdersInProcess, err = meter.Int64UpDownCounter(
		"order_product.pending_orders.in_process",
		metric.WithDescription("Number of pending orders currently being processed"),
		metric.WithUnit("{order}"),
	)
	must(err)

	m.PendingOrdersProcessingDuration, err = meter.Float64Histogram(
		"order_product.pending_orders.processing_duration_seconds",
		metric.WithDescription("Duration of pending order processing"),
		metric.WithUnit("s"),
	)
	must(err)

	m.RecoverProcessedOrdersTotal, err = meter.Int64Counter(
		"order_product.recover_processed_orders.total",
		metric.WithDescription("Total number of recovered processed orders"),
		metric.WithUnit("{order}"),
	)
	must(err)

	m.RecoverProcessedOrdersFailedTotal, err = meter.Int64Counter(
		"order_product.recover_processed_orders.failed_total",
		metric.WithDescription("Number of failed attempts to recover processed orders"),
		metric.WithUnit("{order}"),
	)
	must(err)

	return m
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("failed to create metric: %v", err))
	}
}
