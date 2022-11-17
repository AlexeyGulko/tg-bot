package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	jaggerConfig "github.com/uber/jaeger-client-go/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/config"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

func initConfig() *config.Service {
	cfg, err := config.New()
	if err != nil {
		log.Fatal("config init failed:", err)
	}

	return cfg
}

func initRDB(ctx context.Context, cfg *config.Service) *redis.Client {
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     cfg.RedisHostPort(),
			Password: cfg.RedisPassword(),
			DB:       cfg.RedisDB(),
		})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logger.Fatal("redis init failed:", zap.Error(err))
	}

	return rdb
}

func iniTracing(cfg *config.Service) {
	cfgJagger := jaggerConfig.Configuration{
		Sampler: &jaggerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaggerConfig.ReporterConfig{
			LocalAgentHostPort: cfg.JaegerHostPort(),
		},
	}

	_, err := cfgJagger.InitGlobalTracer(cfg.ServiceName())
	if err != nil {
		logger.Error("tracing init failed:", zap.Error(err))
	}
}

func initDb(cfg *config.Service) *sql.DB {
	dbconfig := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost(),
		cfg.DBPort(),
		cfg.DBUser(),
		cfg.DBPassword(),
		cfg.DBName(),
	)

	db, err := sql.Open("postgres", dbconfig)
	if err != nil {
		logger.Fatal("db init failed:", zap.Error(err))
	}

	return db
}
