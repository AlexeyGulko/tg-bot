package user

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
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
