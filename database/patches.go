package database // import "miniflux.app/database"

import (
	"database/sql"
)

type patcher struct {
	db *sql.DB
}

func (p *patcher) do() error {
	// CREATE TABLE IF NOT EXISTS since Postgres 9.1
	_, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS medias (
			id bigserial not null,
			url text not null,
			url_hash text not null unique,
			mime_type text default '',
			content bytea default E''::bytea,
			size int8 default 0,
			cached bool default 'f',
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
	if !p.columnExists("feeds", "cache_media") {
		_, err = p.db.Exec("alter table feeds add column cache_media bool default 'f';")
		if err != nil {
			return err
		}
	}
	if !p.columnExists("users", "view") {
		_, err = p.db.Exec("alter table users add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	if !p.columnExists("categories", "view") {
		_, err = p.db.Exec("alter table categories add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	if !p.columnExists("feeds", "view") {
		_, err = p.db.Exec("alter table feeds add column view text default 'default';")
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *patcher) columnExists(table string, column string) bool {
	var result int
	query := `SELECT 1 
		FROM information_schema.columns 
		WHERE table_name=$1 and column_name=$2;`
	p.db.QueryRow(query, table, column).Scan(&result)
	return result == 1
}
