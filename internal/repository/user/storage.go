package user

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/opentracing/opentracing-go"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Add(ctx context.Context, model dto.User) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "add user")
	defer span.Finish()
	query, args, err := addQuery(model)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	return err
}

func (s *Storage) Get(ctx context.Context, tgId int64) (*dto.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get user")
	defer span.Finish()
	query, args, err := getQuery(tgId)
	if err != nil {
		return nil, err
	}

	row := s.db.QueryRowContext(ctx, query, args...)

	user, err := userFactory(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Storage) Update(ctx context.Context, user *dto.User) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "update user")
	defer span.Finish()
	builder := getBuilder().Update("users").
		Set("currency", user.Currency).
		Set("updated_at", time.Now()).
		Set("month_budget", user.MonthBudget).
		Where(sq.Eq{"id": user.ID})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, query, args...)

	return err
}

func getBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}

func (s *Storage) GetOrCreate(ctx context.Context, model dto.User) (*dto.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get or create user")
	defer span.Finish()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func(tx *sql.Tx) {
		if err != nil {
			err = tx.Rollback()
		}
	}(tx)

	user, err := GetUserWithTx(ctx, tx, model)

	if err != sql.ErrNoRows && err != nil {
		return nil, err
	}

	if err == sql.ErrNoRows {
		query, args, err := addQuery(model)
		if err != nil {
			return nil, err
		}

		_, err = tx.ExecContext(ctx, query, args...)

		if err != nil {
			return nil, err
		}

		user, err = GetUserWithTx(ctx, tx, model)

		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func GetUserWithTx(ctx context.Context, tx *sql.Tx, model dto.User) (*dto.User, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get user with tx")
	defer span.Finish()
	query, args, err := getQueryByDto(model)
	if err != nil {
		return nil, err
	}

	user, err := userFactory(tx.QueryRowContext(ctx, query, args...))
	if err != nil {
		return nil, err
	}
	return user, nil
}
