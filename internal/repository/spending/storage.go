package spending

import (
	"context"
	"database/sql"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Add(ctx context.Context, model dto.Spending) error {
	builder := getBuilder().Insert("spendings").Columns(
		"created_at",
		"user_id",
		"category",
		"amount",
		"date",
		"id",
	).Values(
		time.Now(),
		model.UserID,
		model.Category,
		model.Amount,
		model.Date,
		uuid.New(),
	)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	return err
}

func (s *Storage) GetReportByCategory(ctx context.Context, UserID uuid.UUID, date time.Time) (map[string][]dto.Spending, error) {

	builder := getBuilder().Select(
		"id",
		"user_id",
		"category",
		"amount",
		"date",
		"created_at",
		"updated_at",
	).From("spendings").Where(sq.Eq{"user_id": UserID}).Where(sq.GtOrEq{"date": date})

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}(rows)

	res := make(map[string][]dto.Spending)

	for rows.Next() {
		var spending dto.Spending
		var updated pq.NullTime
		err = rows.Scan(
			&spending.ID,
			&spending.UserID,
			&spending.Category,
			&spending.Amount,
			&spending.Date,
			&spending.Created,
			&updated,
		)
		spending.Updated = updated.Time
		res[spending.Category] = append(res[spending.Category], spending)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (s *Storage) GetSpendingAmount(
	ctx context.Context,
	UserID uuid.UUID,
	start time.Time,
	end time.Time,
) (decimal.Decimal, error) {
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
