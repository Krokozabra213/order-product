package worker

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/GargantuaLabs/order-product/internal/config"
	"github.com/GargantuaLabs/order-product/internal/service"
)

type Workers struct {
	log     *slog.Logger
	cfg     *config.Config
	service *service.Service
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	svc *service.Service,
) *Workers {

	return &Workers{
		log:     log.With(slog.String("component", "workers")),
		cfg:     cfg,
		service: svc,
	}
}

func (w *Workers) StartPendingOrdersWorker(ctx context.Context) {
	interval := w.cfg.Workers.PendingOrdersInterval
	batchSize := w.cfg.Workers.PendingOrdersBatchSize

	initialJitter := time.Duration(rand.Int63n(int64(interval)))
	w.log.Info("starting pending orders worker", slog.Duration("initial_jitter", initialJitter))
	time.Sleep(initialJitter)

	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		var err error

		for {
			select {
			case <-ctx.Done():
				w.log.Info("pending orders worker stopped")
				return

			case <-ticker.C:
				execJitter := time.Duration(rand.Int63n(int64(interval) / 5))
				time.Sleep(execJitter)

				err = w.service.RunPendingOrdersWorker(ctx, batchSize)
				if err != nil {
					w.log.Error("pending orders worker failed", slog.String("error", err.Error()))
				}
			}
		}
	}()
}

func (w *Workers) StartRecoveryWorker(ctx context.Context) {

	interval := w.cfg.Workers.RecoveryOrderInterval
	timeout := w.cfg.Workers.ProcessingOrderTimeout
	batchSize := w.cfg.Workers.RecoveryOrderBatchSize

	initialJitter := time.Duration(rand.Int63n(int64(interval)))
	w.log.Info("starting recovery worker", slog.Duration("initial_jitter", initialJitter))
	time.Sleep(initialJitter)
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()
		var err error

		for {
			select {
			case <-ctx.Done():
				w.log.Info("recovery worker stopped")
				return

			case <-ticker.C:
				_, err = w.service.RunRecoverStaleProcessingOrders(ctx, timeout, batchSize)
				if err != nil {
					w.log.Error("recovery worker failed", slog.String("error", err.Error()))
				}
			}
		}
	}()
}

func jitter(rng *rand.Rand, interval time.Duration) time.Duration {
	if interval <= 0 {
		return 0
	}
	half := interval / 2
	return time.Duration(rng.Int63n(int64(half*2))) - half
}
