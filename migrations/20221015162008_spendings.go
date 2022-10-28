package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upSpendings, downSpendings)
}

func upSpendings(tx *sql.Tx) error {
	const query = `
	create table spendings
	(
		id integer generated always as identity,
		user_id       integer,
		category    text,
		amount numeric,
		date timestamp,
		created_at timestamp,
		updated_at timestamp
	);
	
	create index user_id_idx on spendings(user_id);
	`

	//Btree user_id тк траты получаются по по пользователю

	_, err := tx.Exec(query)
	return err
}

func downSpendings(tx *sql.Tx) error {
	const query = `
	drop index user_id_idx;
	drop table spendings;
	`

	_, err := tx.Exec(query)
	return err
}
