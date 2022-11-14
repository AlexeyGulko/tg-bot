package spending

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Add(ctx context.Context, model *dto.Spending) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "store spending")
	defer span.Finish()

	id := uuid.New()
	now := time.Now()
	builder := getBuilder().Insert("spendings").Columns(
		"created_at",
		"user_id",
		"category",
		"amount",
		"date",
		"id",
	).Values(
		now,
		model.UserID,
		model.Category,
		model.Amount,
		model.Date,
		id,
	)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	model.ID = id
	model.Created = now

	return nil
}

func (s *Storage) GetReportByCategory(
	ctx context.Context,
	UserID uuid.UUID,
	start time.Time,
	end time.Time,
) (dto.SpendingReport, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get spending report")
	defer span.Finish()

	builder := getBuilder().Select(
		"s.category",
		"s.amount",
		"s.date",
		"u.currency",
		"r.rate",
	).
		LeftJoin("users u on s.user_id = u.id").
		LeftJoin("rates r on r.id = (select rl.id from rates rl where u.currency = rl.code and s.date = rl.ts limit 1)").
		From("spendings s").
		Where(sq.Eq{"s.user_id": UserID}).
		Where(sq.GtOrEq{"s.date": start}).
		Where(sq.LtOrEq{"s.date": end})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	res := make(dto.SpendingReport)

	for rows.Next() {
		var item dto.SpendingReportItem
		var rate decimal.NullDecimal
		err = rows.Scan(
			&item.Category,
			&item.Amount,
			&item.Date,
			&item.Currency,
			&rate,
		)

		item.Rate = rate.Decimal
		res[item.Category] = append(res[item.Category], item)
		if err != nil {
			return nil, err
		}
	}

	return res, err
}

func (s *Storage) GetSpendingAmount(
	ctx context.Context,
	UserID uuid.UUID,
	start time.Time,
	end time.Time,
) (decimal.Decimal, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get spending amount")
	defer span.Finish()
	builder := getBuilder().
		Select("sum(amount)").
		From("spendings").
		Where(sq.GtOrEq{"date": start}).
		Where(sq.LtOrEq{"date": end}).
		Where(sq.Eq{"user_id": UserID})

	query, args, err := builder.ToSql()
	if err != nil {
		return decimal.Decimal{}, err
	}

	row := s.db.QueryRowContext(ctx, query, args...)
	var amount sql.NullFloat64
	err = row.Scan(&amount)
	if err != nil {
		return decimal.Decimal{}, err
	}

	return decimal.NewFromFloat(amount.Float64), nil
}

func getBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
