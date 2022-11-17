package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/lib/pq"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/common/infrastructure/cache"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/repository/spending"
	"go.uber.org/zap"
)

var (
	spendingStorage *spending.StorageWithCache
	client          *ReporterClient
	cfg             *config.Service
)

func main() {
	ctx, closeCtx := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer closeCtx()

	cfg = initConfig()
	logger.Init(cfg)
	iniTracing(cfg)
	db := initDb(cfg)
	rdb := initRDB(ctx, cfg)
	cacheSrv := cache.New(rdb)

	BrokersList := []string{cfg.KafkaHost() + ":" + strconv.Itoa(int(cfg.KafkaPort()))}
	consumerGroupHandler := Consumer{handler: NewHandler(cfg)}
	go func() {
		if err := startConsumerGroup(ctx, BrokersList, consumerGroupHandler); err != nil {
			logger.Fatal("consuming init failed:", zap.Error(err))
		}
	}()

	spendingStorage = spending.NewStorageWithCache(db, cacheSrv)

	grpcClient := NewClient(cfg)

	grpcClient.startGrpcClient(ctx)

	client = NewReporterClient(grpcClient)

	logger.Info("reporter started")

	<-ctx.Done()
	logger.Info("reporter stopped")
}
