package spending

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
)

var ttl = time.Duration(0)

type StorageWithCache struct {
	Storage
	cache Cache
}

type Cache interface {
	Add(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, val interface{}) (bool, error)
	DeleteByKeyPart(ctx context.Context, key string) error
}

func NewStorageWithCache(db *sql.DB, cache Cache) *StorageWithCache {
	return &StorageWithCache{cache: cache, Storage: Storage{db: db}}
}

func (s *StorageWithCache) Add(ctx context.Context, model *dto.Spending) error {
	if err := s.Storage.Add(ctx, model); err != nil {
		return err
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "delete reports from cache")
	defer span.Finish()

	key := fmt.Sprintf("report_%s", model.UserID)

	if err := s.cache.DeleteByKeyPart(ctx, key); err != nil {
		return nil
	}

	return nil
}

func (s *StorageWithCache) GetReportByCategory(
	ctx context.Context,
	user *dto.User,
	start time.Time,
	end time.Time,
) (dto.SpendingReport, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get spending report from cache")
	defer span.Finish()

	key := fmt.Sprintf("report_%s_%s_%d_%d", user.ID, user.Currency, start.Unix(), end.Unix())
	var spending dto.SpendingReport
	has, err := s.cache.Get(ctx, key, &spending)
	if err != nil {
		return nil, err
	}

	if has {
		return spending, nil
	}

	if spending, err = s.Storage.GetReportByCategory(ctx, user.ID, start, end); err != nil {
		return nil, err
	}

	if err := s.cache.Add(ctx, key, spending, ttl); err != nil {
		return nil, err
	}

	return spending, nil
}
