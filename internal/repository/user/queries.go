package user

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

func getQuery(tgId int64) (string, []interface{}, error) {
	builder := getBuilder().Select(
		"id",
		"tg_id",
		"currency",
		"created_at",
		"updated_at",
		"month_budget",
	).From("users").Where(sq.Eq{"tg_id": tgId})

	query, args, err := builder.ToSql()
	if err != nil {
		return query, args, err
	}

	return query, args, nil
}

func getQueryByDto(dto dto.User) (string, []interface{}, error) {
	if dto.TgID == 0 && dto.ID.ID() == 0 {
		return "", nil, errors.New("empty dto")
	}
	builder := getBuilder().Select(
		"id",
		"tg_id",
		"currency",
		"created_at",
		"updated_at",
		"month_budget",
	).From("users")

	if dto.TgID != 0 {
		builder = builder.Where(sq.Eq{"tg_id": dto.TgID})
	}
	if dto.ID.ID() != 0 {
		builder = builder.Where(sq.Eq{"id": dto.ID})
	}

	query, args, err := builder.ToSql()
	if err != nil {
		return query, args, err
	}

	return query, args, nil
}

func addQuery(user dto.User) (string, []interface{}, error) {
	builder := getBuilder().Insert("users").Columns(
		"created_at",
		"tg_id",
		"currency",
		"id",
	).Values(
		time.Now(),
		user.TgID,
		user.Currency,
		uuid.New(),
	)
	query, args, err := builder.ToSql()
	if err != nil {
		return query, args, err
	}

	return query, args, nil
}
