package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(upUuid, downUuid)
}

func upUuid(tx *sql.Tx) error {
	const query = `
		truncate users, spendings, rates;
		alter table users drop column id;
		alter table users add column id uuid;
		alter table users add primary key (id);
		alter table rates drop column id;
		alter table rates add column id uuid;
		alter table rates add primary key (id);
		alter table spendings drop column id;
		alter table spendings add column id uuid;
		alter table spendings add primary key (id);
		alter table spendings drop column user_id;
		alter table spendings add column user_id uuid;
	`

	_, err := tx.Exec(query)
	return err
}

func downUuid(tx *sql.Tx) error {
	const query = `
		alter table users drop constraint users_pkey;
		alter table users drop column id;
		alter table users add column id integer;
		alter table users alter column id set not null;
		alter table users alter column id add generated always as identity ;
		alter table rates drop constraint rates_pkey;
		alter table rates drop column id;
		alter table rates add column id integer;
		alter table rates alter column id set not null;
		alter table rates alter column id add generated always as identity ;
		alter table spendings drop constraint spendings_pkey;
		alter table spendings drop column id;
		alter table spendings add column id integer;
		alter table spendings alter column id set not null;
		alter table spendings alter column id add generated always as identity ;
		alter table spendings drop column user_id;
		alter table spendings add column user_id integer;
	`

	_, err := tx.Exec(query)
	return err
}
