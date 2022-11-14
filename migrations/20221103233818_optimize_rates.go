package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upOptimizeRates, downOptimizeRates)
}

func upOptimizeRates(tx *sql.Tx) error {
	// This code is executed when the migration is applied.

	const query = `
	create unique index rate_code_ts_unique_idx
	on rates (code, ts);
	alter table rates drop column created_at;
	alter table rates drop column updated_at;
	`

	_, err := tx.Exec(query)
	return err
}

func downOptimizeRates(tx *sql.Tx) error {
	const query = `
	drop index rate_code_ts_unique_idx;
	alter table rates add column created_at timestamptz;
	alter table rates add column updated_at timestamptz;
	`

	_, err := tx.Exec(query)
	return err
}
