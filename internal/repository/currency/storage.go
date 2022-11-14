package currency

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
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
	currency.ID = uuid.New()
	builder := getBuilder().Insert("rates").Columns(
		"code",
		"rate",
		"ts",
		"id",
	).Values(
		currency.Code,
		currency.Rate,
		currency.TimeStamp,
		currency.ID,
	)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	if _, err = s.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

func (s *Storage) AddBulk(ctx context.Context, currencies []*dto.Currency) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "store currency bulk")
	defer span.Finish()
	builder := getBuilder().Insert("rates").Columns(
		"code",
		"rate",
		"ts",
		"id",
	)

	for _, curr := range currencies {
		curr.ID = uuid.New()
		builder = builder.Values(
			curr.Code,
			curr.Rate,
			curr.TimeStamp,
			curr.ID,
		)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	if _, err = s.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, date time.Time, code string) (*dto.Currency, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get currency")
	defer span.Finish()

	var currency dto.Currency

	builder := getBuilder().Select(
		"id",
		"code",
		"rate",
		"ts",
	).From("rates").Where(sq.Eq{"code": code})

	if !date.IsZero() {
		builder = builder.Where(sq.Eq{"ts": date})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, query, args...)

	err = row.Scan(
		&currency.ID,
		&currency.Code,
		&currency.Rate,
		&currency.TimeStamp,
	)

	if err != nil {
		return nil, err
	}

	return &currency, nil
}

func getBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
