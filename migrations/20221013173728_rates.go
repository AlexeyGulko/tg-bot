package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upRates, downRates)
}

func upRates(tx *sql.Tx) error {
	const query = `
	create table rates
	(
		id integer generated always as identity,
		code       text,
		rate    bigint,
		ts         timestamp,
		created_at timestamp,
		updated_at timestamp
	);
	
	create index rates_code_ts_idx on rates(code, ts);
	`
	//копипаста с воркшопа

	_, err := tx.Exec(query)
	return err

}

func downRates(tx *sql.Tx) error {
	const query = `
	drop index rates_code_ts_idx;
	drop table rates;
	`

	_, err := tx.Exec(query)
	return err

}
