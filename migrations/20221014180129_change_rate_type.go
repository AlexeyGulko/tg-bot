package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upChangeRateType, downChangeRateType)
}

func upChangeRateType(tx *sql.Tx) error {
	const query = `alter table rates alter column rate type numeric;`

	_, err := tx.Exec(query)
	return err
}

func downChangeRateType(tx *sql.Tx) error {
	const query = `alter table rates alter column rate type bigint;`

	_, err := tx.Exec(query)
	return err
}
