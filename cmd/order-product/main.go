package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	commonclient "github.com/Krokozabra213/gargantua_common/pkg/clients"
	"github.com/Krokozabra213/gargantua_common/pkg/logger"
	"github.com/GargantuaLabs/order-product/internal/client"
	"github.com/GargantuaLabs/order-product/internal/config"
	infrapg "github.com/GargantuaLabs/order-product/internal/infrastructure/postgres"
	orderRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/order"
	productRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/product"
	subRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/subscription"
	subActiveRepo "github.com/GargantuaLabs/order-product/internal/repository/postgres/subscription_activation"
	redisRepo "github.com/GargantuaLabs/order-product/internal/repository/redis_cache"
	"github.com/GargantuaLabs/order-product/internal/server/grpc/app"
	"github.com/GargantuaLabs/order-product/internal/server/grpc/handlers"
	"github.com/GargantuaLabs/order-product/internal/service"
	"github.com/GargantuaLabs/order-product/internal/worker"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Init()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	telemetry, err := setupTelemetry(ctx, cfg)
	if err != nil {
		return err
	}
	defer telemetry.Shutdown(ctx)

	log := logger.Init(&cfg.Logger, telemetry.Handler)

	pool, err := commonclient.NewPostgresClient(&cfg.Postgres)
	if err != nil {
		return err
	}
	defer pool.Close()

	redisClient, err := commonclient.NewRedisClient(&cfg.Redis)
	if err != nil {
		panic(err)
	}

	_ = redisClient

	productRepo := productRepo.New(pool)
	orderRepo := orderRepo.New(pool)
	subRepo := subRepo.New(pool)
	actRepo := subActiveRepo.New(pool)
	cacheRepo := redisRepo.New(redisClient, cfg.Redis.Prefix)

	balanceClient, err := client.NewBalance(ctx, cfg.Balance.Addr)
	if err != nil {
		return err
	}
	defer func() {
		if err := balanceClient.Close(); err != nil {
			log.Error("ошибка при закрытии balanceClient: %v", "error", err)
		}
	}()

	uow := infrapg.NewUnitOfWork(pool)
	providers := service.NewProviders(cacheRepo, productRepo, orderRepo, subRepo, actRepo)

	service := service.New(uow, cfg, log, providers, balanceClient)

	handler := handlers.New(cfg, service)

	grpcServer := app.New(&cfg.GRPC, log, handler)
	go grpcServer.MustRun()

	worker := worker.New(log, cfg, service)
	worker.StartPendingOrdersWorker(ctx)
	worker.StartRecoveryWorker(ctx)

	<-ctx.Done()

	log.Info("shutting down application")

	grpcServer.Stop()

	log.Info("application stopped")

	return nil
}
