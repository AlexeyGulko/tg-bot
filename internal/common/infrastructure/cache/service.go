package cache

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

type Service struct {
	db *redis.Client
}

func New(db *redis.Client) *Service {
	return &Service{db: db}
}

func (s *Service) Add(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	var marshed []byte
	var err error
	if marshed, err = json.Marshal(val); err != nil {
		return err
	}

	if err := s.db.Set(ctx, key, marshed, ttl).Err(); err != nil {
		return err
	}
	logger.Info("cached", zap.String("key", key))

	return nil
}

func (s *Service) Get(ctx context.Context, key string, val interface{}) (bool, error) {
	cached, err := s.db.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return false, err
	}

	if err == redis.Nil {
		return false, nil
	}

	if err := json.Unmarshal(cached, &val); err != nil {
		return false, err
	}

	logger.Info("cache hit", zap.String("key", key))
	return true, nil
}

func (s *Service) DeleteByKeyPart(ctx context.Context, key string) error {
	iter := s.db.Scan(ctx, 0, key+"*", 0).Iterator()

	pipe := s.db.Pipeline()
	for iter.Next(ctx) {
		if err := pipe.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	if err := iter.Err(); err != nil {
		return err
	}

	logger.Info("cache delete", zap.String("key", key))
	return nil
}
