package currency

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
	"golang.org/x/net/context"
)

var ttl = time.Minute * 1

type StorageWithCache struct {
	Storage
	cache Cache
}

type Cache interface {
	Add(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string, val interface{}) (bool, error)
}

func NewStorageWithCache(db *sql.DB, rdb Cache) *StorageWithCache {
	return &StorageWithCache{cache: rdb, Storage: Storage{db: db}}
}

func (c *StorageWithCache) Add(ctx context.Context, currency dto.Currency) error {
	if err := c.Storage.Add(ctx, currency); err != nil {
		return err
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "add currency to cache")
	defer span.Finish()

	key := fmt.Sprintf("currency_%s_%s", currency.Code, currency.TimeStamp)
	if err := c.cache.Add(ctx, key, currency, ttl); err != nil {
		return err
	}

	return nil
}

func (c *StorageWithCache) AddBulk(ctx context.Context, currencies []*dto.Currency) error {
	if err := c.Storage.AddBulk(ctx, currencies); err != nil {
		return nil
	}

	span, ctx := opentracing.StartSpanFromContext(ctx, "add bulk currency to cache")
	defer span.Finish()

	for _, curr := range currencies {
		key := fmt.Sprintf("currency_%s_%d", curr.Code, curr.TimeStamp.Unix())
		if err := c.cache.Add(ctx, key, curr, ttl); err != nil {
			return err
		}
	}

	return nil
}

func (c *StorageWithCache) Get(ctx context.Context, date time.Time, code string) (*dto.Currency, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get currency from cache")
	defer span.Finish()

	var currency *dto.Currency
	key := fmt.Sprintf("currency_%s_%d", code, date.Unix())
	cached, err := c.cache.Get(ctx, key, &currency)
	if err != nil {
		return nil, err
	}

	if cached {
		return currency, nil
	}

	currency, err = c.Storage.Get(ctx, date, code)
	if err != nil {
		return nil, err
	}

	key = fmt.Sprintf("currency_%s_%d", currency.Code, currency.TimeStamp.Unix())

	if err := c.cache.Add(ctx, key, currency, 0); err != nil {
		return currency, err
	}

	return currency, nil
}
