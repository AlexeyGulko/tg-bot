package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upUser, downUser)
}

func upUser(tx *sql.Tx) error {
	const query = `
	create table users
	(
		id integer generated always as identity,
		currency       text,
		tg_id    integer,
		created_at timestamp,
		updated_at timestamp
	);
	
	create index tg_id_idx on users(tg_id);
	`
	//Btree для tg_id, используется для получения пользователя
	_, err := tx.Exec(query)
	return err

}

func downUser(tx *sql.Tx) error {
	const query = `
	drop index tg_id_idx;
	drop table users;
	`

	_, err := tx.Exec(query)
	return err

}
