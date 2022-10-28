package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upUserMonthBudget, downUserMonthBudget)
}

func upUserMonthBudget(tx *sql.Tx) error {
	const query = `
		alter table users add column month_budget numeric;
		create index users_budget_idx on users(month_budget)
	`

	_, err := tx.Exec(query)
	return err
}

func downUserMonthBudget(tx *sql.Tx) error {
	const query = `
		drop index users_budget_idx;
		alter table users drop column month_budget;
	`

	_, err := tx.Exec(query)
	return err
}
