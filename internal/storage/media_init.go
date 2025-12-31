package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"log/slog"

	"miniflux.app/v2/internal/model"
)

// CreateMediasRunOnce create media records for starred and unread entries,
// runs once only when the entry_medias table is empty, which could be two use cases:
// First, system has just migrated to media cache feature support.
// Second, database has just restored from backup, since user may not want to include huge medias into backup.
func (s *Storage) CreateMediasRunOnce() {
	has, err := hasMediaRecord(s.db)
	if has || err != nil {
		return
	}
	var startID int64
	tx, err := s.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	for {
		entries, err := getEntriesForCreateMediasRunOnce(s.db, startID)
		if err != nil {
			slog.Error("[Storage:CreateMediasRunOnce] Error: %s", slog.Any("error", err))
			break
		}
		if len(entries) == 0 {
			break
		}
		err = s.CreateEntriesMedia(tx, entries)
		if err != nil {
			slog.Error("[Storage:CreateMediasRunOnce] Error: %s", slog.Any("error", err))
		}
		startID = entries[len(entries)-1].ID
	}
	tx.Commit()
}

// hasMediaRecord returns if entry_medias has at least one record.
func hasMediaRecord(db *sql.DB) (bool, error) {
	var result int
	query := `SELECT count(*) FROM (SELECT * FROM entry_medias LIMIT 1) as m`
	err := db.QueryRow(query).Scan(&result)
	if err != nil {
		return false, err
	}
	return result != 0, nil
}

func getEntriesForCreateMediasRunOnce(db *sql.DB, startID int64) (model.Entries, error) {
	entries := make(model.Entries, 0)
	query := `
		SELECT id, content 
		FROM entries e 
			LEFT JOIN entry_medias em on e.id=em.entry_id
		WHERE id>$1 AND (status<>$2) AND em.entry_id IS NULL
		ORDER BY e.id ASC
		LIMIT 1000
	`
	rows, err := db.Query(query, startID, model.EntryStatusRemoved)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return entries, nil
	} else if err != nil {
		return nil, err
	}
	for rows.Next() {
		var entry model.Entry
		err := rows.Scan(&entry.ID, &entry.Content)
		if err != nil {
			return nil, err
		}
		entries = append(entries, &entry)

	}
	return entries, nil
}
