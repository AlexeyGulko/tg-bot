package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upTz, downTz)
}

func upTz(tx *sql.Tx) error {

	const query = `
	alter table rates
	alter column ts type timestamptz;
	alter table rates
	alter column created_at type timestamptz;
	alter table rates
	alter column updated_at type timestamptz;

	alter table spendings
	alter column updated_at type timestamptz;
	alter table spendings
	alter column created_at type timestamptz;
	alter table spendings
	alter column date type timestamptz;

	alter table users
	alter column updated_at type timestamptz;
	alter table users
	alter column created_at type timestamptz;
	`

	_, err := tx.Exec(query)
	return err
}

func downTz(tx *sql.Tx) error {
	const query = `
	alter table rates
	alter column ts type timestamp;
	alter table rates
	alter column created_at type timestamp;
	alter table rates
	alter column updated_at type timestamp;
	alter table spendings
	alter column updated_at type timestamp;
	alter table spendings
	alter column created_at type timestamp;
	alter table users
	alter column updated_at type timestamp;
	alter table users
	alter column created_at type timestamp;
	alter table spendings
	alter column date type timestamp;
	`

	_, err := tx.Exec(query)
	return err
}
