package currency

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Add(ctx context.Context, currency dto.Currency) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "store currency")
	defer span.Finish()
	builder := getBuilder().Insert("rates").Columns(
		"created_at",
		"code",
		"rate",
		"ts",
		"id",
	).Values(
		time.Now(),
		currency.Code,
		currency.Rate,
		currency.TimeStamp,
		uuid.New(),
	)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	return err
}

func (s *Storage) AddBulk(ctx context.Context, currencies []dto.Currency) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "store currency bulk")
	defer span.Finish()
	builder := getBuilder().Insert("rates").Columns(
		"created_at",
		"code",
		"rate",
		"ts",
		"id",
	)

	for _, curr := range currencies {
		builder = builder.Values(
			time.Now(),
			curr.Code,
			curr.Rate,
			curr.TimeStamp,
			uuid.New(),
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)
	return err
}

func (s *Storage) Get(ctx context.Context, date time.Time, code string) (*dto.Currency, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get currency")
	defer span.Finish()
	builder := getBuilder().Select(
		"id",
		"code",
		"rate",
		"ts",
		"created_at",
		"updated_at",
	).From("rates").Where(sq.Eq{"code": code})

	if !date.IsZero() {
		builder = builder.Where(sq.Eq{"ts": date})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var currency dto.Currency
	var updated pq.NullTime
	row := s.db.QueryRowContext(ctx, query, args...)

	err = row.Scan(
		&currency.ID,
		&currency.Code,
		&currency.Rate,
		&currency.TimeStamp,
		&currency.Created,
		&updated,
	)

	currency.Updated = updated.Time
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

func getBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
