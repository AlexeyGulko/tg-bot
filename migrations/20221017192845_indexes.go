package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upIndexes, downIndexes)
}

func upIndexes(tx *sql.Tx) error {
	const query = `
	create index rates_code_idx on rates(code);
	create index rates_ts_idx on rates using brin (ts);
	create index spendings_date_idx on spendings using brin (date);
	`

	//brin подходит для дат и легковесный, код по b tree для ускорения поиска
	_, err := tx.Exec(query)
	return err
}

func downIndexes(tx *sql.Tx) error {
	const query = `
	drop index rates_code_idx;
	drop index rates_ts_idx;
	drop index spendings_date_idx;
	`

	_, err := tx.Exec(query)
	return err
}
