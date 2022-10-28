package user

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gitlab.ozon.dev/dev.gulkoalexey/gulko-alexey/internal/dto"
)

func userFactory(row *sql.Row) (*dto.User, error) {
	var user dto.User
	var updated pq.NullTime
	var budget sql.NullFloat64

	err := row.Scan(
		&user.ID,
		&user.TgID,
		&user.Currency,
		&user.Created,
		&updated,
		&budget,
	)

	user.MonthBudget = decimal.NewFromFloat(budget.Float64)
	user.Updated = updated.Time
	if err != nil {
		return nil, err
	}
	return &user, nil
}
