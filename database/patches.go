package database // import "miniflux.app/database"

import (
	"database/sql"
)

func patch(tx *sql.Tx) error {
	// CREATE TABLE IF NOT EXISTS since Postgres 9.1
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS medias (
			id bigserial not null,
			url text not null,
			url_hash text not null unique,
			mime_type text default '',
			content bytea default null,
			size int8 default 0,
			cached bool default 'f',
			error_count int default 0,
			created_at timestamp with time zone default current_timestamp,
			primary key (id)
		);
		CREATE TABLE IF NOT EXISTS entry_medias (
			entry_id int8 NOT NULL,
			media_id int8 NOT NULL,
			use_cache bool default 'f',
			PRIMARY KEY (entry_id, media_id),
			foreign key (entry_id) references entries(id) on delete cascade,
			foreign key (media_id) references medias(id) on delete cascade
		);`)
	if err != nil {
		return err
	}
	if !columnExists(tx, "feeds", "cache_media") {
		_, err = tx.Exec("alter table feeds add column cache_media bool default 'f';")
		if err != nil {
			return err
		}
	}
	if !columnExists(tx, "users", "view") {
		_, err = tx.Exec("alter table users add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	if !columnExists(tx, "categories", "view") {
		_, err = tx.Exec("alter table categories add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	if !columnExists(tx, "feeds", "view") {
		_, err = tx.Exec("alter table feeds add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	if !columnExists(tx, "feeds", "nsfw") {
		_, err = tx.Exec("alter table feeds add column nsfw bool default 'f';")
		if err != nil {
			return err
		}
	}
	if !columnExists(tx, "medias", "error_count") {
		_, err = tx.Exec("alter table medias add column error_count int default 0;")
		if err != nil {
			return err
		}
	}
	return nil
}

func columnExists(tx *sql.Tx, table string, column string) bool {
	var result int
	query := `SELECT 1 
		FROM information_schema.columns 
		WHERE table_name=$1 and column_name=$2;`
	tx.QueryRow(query, table, column).Scan(&result)
	return result == 1
}
