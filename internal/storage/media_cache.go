package storage // import "miniflux.app/v2/internal/storage"

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"miniflux.app/v2/internal/filesystem"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/media"
)

const maxCachingError = 3

// CacheMedias caches medias of starred entries.
// caching task has two parts:
// 1. make sure the media is cached
// 2. the entry_medias record claims to use the media, by setting 'use_cache' to true
func (s *Storage) CacheMedias() error {
	offset := 0
	limit := 5000
	count := 0
	failedCount := 0
	defer func() {
		slog.Info("cache medias", slog.Int("cached", count), slog.Int("failed", failedCount))
	}()

	var (
		medias   model.Medias
		mEntries map[int64]string
		err      error
	)
	for flag := true; flag || len(medias) > 0; medias, mEntries, err = s.getUncachedMedias(offset, limit) {
		flag = false
		if err != nil {
			return err
		}
		for _, m := range medias {
			entries, _ := mEntries[m.ID]
			if !m.Cached {
				// try load media from disk cache first
				if err = filesystem.MediaFromCache(m); err != nil {
					slog.Debug(
						"[Storage:CacheMedias] unable to load disk cache: %v",
						slog.Any("error", err),
					)
					if err = media.FindMedia(m); err != nil {
						slog.Error("[Storage:CacheMedias] unable to fetch media %s: %v", m.URL, err)
						m.ErrorCount++
						if err = s.UpdateMediaError(m); err != nil {
							return fmt.Errorf("[Storage:CacheMedias] unable to update media error #%d: %v", m.ID, err)
						}
						// make sure this media not selected again
						// ErrorCount >= maxCachingError are excluded already,
						// we only have to move offset when it is smaller
						if m.ErrorCount < maxCachingError {
							offset++
						}
						failedCount++
						continue
					}
				}
				m.Cached = true
				// reset error count on success
				m.ErrorCount = 0
				if err = s.UpdateMedia(m); err != nil {
					return fmt.Errorf("[Storage:CacheMedias] unable to update media #%d: %v", m.ID, err)
				}
			}
			sql := fmt.Sprintf(`UPDATE entry_medias set use_cache='t' WHERE media_id=%d AND entry_id in (%s)`, m.ID, entries)
			if _, err = s.db.Exec(sql); err != nil {
				return fmt.Errorf("[Storage:CacheMedias] unable to update media references media_id=%d, entry_id=(%s) : %v", m.ID, entries, err)
			}
			count++
		}
	}
	return nil
}

// getUncachedMedias gets medias which should be but not yet cached, together with their entry referrers' IDs
func (s *Storage) getUncachedMedias(offset int, limit int) (model.Medias, map[int64]string, error) {
	mediaEntries := make(map[int64]string, 0)
	query := `
	SELECT 
		m.id,
		m.url,
		m.url_hash,
		m.mime_type,
		m.cached,
		m.error_count,
		max(e.url) as referrer,
		string_agg(cast(e.id as TEXT),',') as eids
    FROM feeds f
        INNER JOIN entries e ON f.id=e.feed_id
        INNER JOIN entry_medias em ON e.id=em.entry_id
        INNER JOIN medias m ON em.media_id=m.id
    WHERE 
        f.cache_media='T' 
        AND e.starred='T' 
		AND (em.use_cache='F' OR m.cached='F')
		AND m.error_count < $1
	GROUP BY m.id
	OFFSET $2
    LIMIT $3
`

	medias := make(model.Medias, 0)
	rows, err := s.db.Query(query, maxCachingError, offset, limit)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, mediaEntries, nil
	} else if err != nil {
		return nil, nil, fmt.Errorf("unable to fetch uncached medias: %v", err)
	}

	for rows.Next() {
		var media model.Media
		var entryIDs string
		err := rows.Scan(
			&media.ID,
			&media.URL,
			&media.URLHash,
			&media.MimeType,
			&media.Cached,
			&media.ErrorCount,
			&media.Referrer,
			&entryIDs,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to fetch uncached medias row: %v", err)
		}
		medias = append(medias, &media)
		mediaEntries[media.ID] = entryIDs

	}
	return medias, mediaEntries, nil
}

// HasEntryCache indicates if an entry has cache.
func (s *Storage) HasEntryCache(entryID int64) bool {
	var result bool
	query := `
		SELECT true
		FROM entry_medias em
			INNER JOIN medias m on em.media_id=m.id
		WHERE em.entry_id=$1 AND em.use_cache='T' AND m.cached='T'
	`
	s.db.QueryRow(query, entryID).Scan(&result)
	return result
}

// CacheEntryMedias caches media of an entry.
func (s *Storage) CacheEntryMedias(userID, EntryID int64) error {
	medias, err := s.getEntryMedias(userID, EntryID)
	if err != nil {
		return err
	}
	if len(medias) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for _, m := range medias {
		if !m.Cached {
			// try load media from disk cache first
			if err = filesystem.MediaFromCache(m); err != nil {
				slog.Debug(
					"[Storage:CacheEntryMedias] unable to load disk cache",
					slog.Any("error", err),
				)
				if err = media.FindMedia(m); err != nil {
					slog.Error("[Storage:CacheEntryMedias] unable to fetch media %s: %v", m.URL, err)
					m.ErrorCount++
					if err = s.UpdateMediaError(m); err != nil {
						return fmt.Errorf("[Storage:CacheEntryMedias] unable to update media error #%d: %v", m.ID, err)
					}
					continue
				}
			}
			m.Cached = true
			m.ErrorCount = 0
			if err = s.UpdateMedia(m); err != nil {
				return fmt.Errorf("[Storage:CacheEntryMedias] unable to update media #%d: %v", m.ID, err)
			}
		}
		buf.WriteString(fmt.Sprintf("('%v','%v','T'),", EntryID, m.ID))
	}
	vals := buf.String()[:buf.Len()-1]
	sql := fmt.Sprintf(`
		INSERT INTO entry_medias (entry_id, media_id, use_cache)
		VALUES %s
		ON CONFLICT (entry_id, media_id) DO UPDATE
			SET use_cache='T'
	`, vals)
	_, err = s.db.Exec(sql)
	if err != nil {
		return fmt.Errorf("[Storage:CacheEntryMedias] unable to update cache references: %v", err)
	}
	return nil
}

func (s *Storage) getEntryMedias(userID, EntryID int64) (model.Medias, error) {
	query := `
		SELECT m.id, m.url, m.url_hash, m.mime_type , m.cached, e.url
		FROM feeds f
			INNER JOIN entries e on f.id=e.feed_id
			INNER JOIN entry_medias em on e.id=em.entry_id
			INNER JOIN medias m on m.id=em.media_id
		WHERE f.user_id=$1 AND e.id=$2 AND m.error_count < $3
`
	medias := make(model.Medias, 0)
	rows, err := s.db.Query(query, userID, EntryID, maxCachingError)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch entry medias: %v", err)
	}

	for rows.Next() {
		var media model.Media
		err := rows.Scan(&media.ID, &media.URL, &media.URLHash, &media.MimeType, &media.Cached, &media.Referrer)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry medias row: %v", err)
		}
		medias = append(medias, &media)

	}
	return medias, nil
}

// RemoveFeedCaches update all cache references of a feed, to not claim to use the cache of media.
// It doesn't really remove the caches in database or disk.
// Unclaimed caches will be remove by CleanMediaCaches() later.
func (s *Storage) RemoveFeedCaches(userID, feedID int64) error {
	result, err := s.db.Exec(`
		UPDATE entry_medias 
		SET use_cache ='f'
		WHERE entry_id in (
			SELECT em.entry_id
            FROM feeds f
                INNER JOIN entries e on f.id=e.feed_id
                INNER JOIN entry_medias em ON e.id=em.entry_id
			WHERE f.id=$1 AND f.user_id=$2
		)
	`, feedID, userID)
	if err != nil {
		return fmt.Errorf("unable to remove cache for feed #%d: %v", feedID, err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to remove cache for feed #%d: %v", feedID, err)
	}

	if count == 0 {
		return errors.New("no cache has been removed")
	}

	return nil
}

// CleanupMedia removes media that no entry claims to use.
func (s *Storage) CleanupMedia() error {
	// Step 1: clean media which has no 'use cache' reference, which applies to 2 cases:
	// 1. media which has reference records, but no 'use cache' record.
	// 2. media which has no reference record at all.
	// After this step, all unused cache content (in database and disk) are removed
	query := `
		UPDATE medias 
		SET content = NULL, cached='f'
		WHERE id in (
			SELECT id
			FROM medias
			WHERE cached='t' and id NOT IN(
				SELECT media_id from entry_medias WHERE use_cache='t'
			)
		)
		RETURNING url_hash
	`
	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("unable to clean up caches: %v", err)
	}

	count := 0
	for rows.Next() {
		var urlHash string
		err := rows.Scan(&urlHash)
		if err != nil {
			slog.Error(
				"unable to fetch unused cache info",
				slog.Any("error", err),
			)
			continue
		}
		err = filesystem.RemoveMediaFile(urlHash)
		if err != nil {
			slog.Error(
				"unable to remove cache file (%s)", urlHash,
				slog.Any("error", err),
			)
			continue
		}
		count++
	}

	slog.Info("remove unused media cache.", slog.Int("count", count))

	// step 2: Remove media records which has no reference record at all.
	err = s.cleanMediaReferences()
	if err != nil {
		return fmt.Errorf(`store: unable to clean media references: %v`, err)
	}
	err = s.cleanMediaRecords()
	if err != nil {
		return fmt.Errorf(`store: unable to clean media records: %v`, err)
	}
	return nil
}

// ToggleEntryCache toggles entry cache.
// Toggle off: update all cache references of the entry, to not claim to use the cache of media.
// It doesn't really remove the caches in database or disk.
// Unclaimed caches will be remove by CleanMediaCaches() later.
// Toggle on: Cache media and set use_cache of cache references to true
// Usually, since toggle off doesn't really remove the caches,
// it just update the use_cache flag to true with very low cost
func (s *Storage) ToggleEntryCache(userID int64, entryID int64) error {
	if s.HasEntryCache(entryID) {
		query := `
			UPDATE entry_medias SET use_cache='f' WHERE entry_id in (
				SELECT e.id
				FROM feeds f
					INNER JOIN entries e on f.id=e.feed_id
				WHERE f.user_id=$1 AND e.id=$2
			);
		`
		result, err := s.db.Exec(query, userID, entryID)
		if err != nil {
			return fmt.Errorf("unable to toggle cache for user #%d, entry #%d: %v", userID, entryID, err)
		}

		count, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("unable to toogle cache for user #%d, entry #%d: %v", userID, entryID, err)
		}
		if count == 0 {
			return errors.New("nothing has been updated")
		}
		return err
	}

	return s.CacheEntryMedias(userID, entryID)
}

// MoveCacheToDisk move all caches in database to disk
func (s *Storage) MoveCacheToDisk() error {
	query := `
		SELECT id, url_hash, content
		FROM medias
		WHERE cached='t' and content IS NOT NULL
	`
	rows, err := s.db.Query(query)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("unable to move cache: %v", err)
	}

	var media model.Media
	var countSaved int64
	for rows.Next() {
		err := rows.Scan(&media.ID, &media.URLHash, &media.Content)
		if err != nil {
			return fmt.Errorf("unable to fetch media row: %v", err)
		}
		err = filesystem.SaveMediaFile(&media)
		if err != nil {
			return fmt.Errorf("unable to save media file: %v", err)
		}
		countSaved++
	}
	query = `
		UPDATE medias 
		SET content=NULL
		WHERE content IS NOT NULL
	`
	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("unable to clear data base cache: %v", err)
	}
	countAll, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to clear data base cache: %v", err)
	}
	slog.Info("cache moved to disk", slog.Int64("count", countSaved), slog.Int64("removed", countAll-countSaved))
	return nil
}

// ValidateCaches finds missing caches and tags cached to false
func (s *Storage) ValidateCaches() error {
	var count int64
	defer func() {
		if count > 0 {
			slog.Info("invalid caches found ", slog.Int64("count", count))
			return
		}
		// slog.Info("caches are good :)")
	}()
	var offset int64
	var limit int64 = 5000
	for {
		// we only have to validate caches on disk
		rows, err := s.db.Query(`
			SELECT id, url_hash
			FROM medias 
			WHERE cached='t' AND content IS NULL
			OFFSET $1
			LIMIT $2
		`, offset, limit)
		// doesn't always returl sql.ErrNoRows when no rows
		if err == sql.ErrNoRows || !rows.Next() {
			return nil
		} else if err != nil {
			return fmt.Errorf("unable to fetch cached media: %v", err)
		}

		var invalidIDsBuf bytes.Buffer
		for {
			var id int64
			var urlHash string
			err := rows.Scan(&id, &urlHash)
			if err != nil {
				return fmt.Errorf("unable to fetch cache info: %v", err)
			}
			exists, err := filesystem.ExistMediaFile(urlHash)
			if err != nil {
				return fmt.Errorf("unable to validate disk cached: %v", err)
			}
			if !exists {
				invalidIDsBuf.WriteString(strconv.Itoa(int(id)))
				invalidIDsBuf.WriteByte(',')
			}
			if !rows.Next() {
				break
			}
		}
		rows.Close()
		offset += limit
		if invalidIDsBuf.Len() > 0 {
			invalidIDs := invalidIDsBuf.String()[:invalidIDsBuf.Len()-1]
			query := fmt.Sprintf(`UPDATE medias SET cached='f' WHERE id in (%s)`, invalidIDs)
			result, err := s.db.Exec(query)
			if err != nil {
				return fmt.Errorf("Unable to update media table for invalid caches: %v", err)
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return fmt.Errorf("Unable to update media table for invalid caches: %v", err)
			}
			count += affected
			offset -= affected
		}
	}
}

// CachedMediasStat gets the count and size of cached media from the database.
func (s *Storage) CachedMediasStat() (count int64, size int64, maxID int64, err error) {
	query := `
		SELECT
			COUNT(1) cnt,
			SUM(m.size) size,
			MAX(m.id) max_id
		FROM medias m
		WHERE m.cached='T'
	`
	err = s.db.QueryRow(query).Scan(&count, &size, &maxID)
	return count, size, maxID, err
}

// CachedMedias gets all cached media from the database.
func (s *Storage) CachedMedias(offset int64, limit int64) (map[string]model.Medias, error) {
	query := `
	SELECT
		MAX(f.title) as feed,
		m.id,
		m.url,
		m.url_hash,
		m.mime_type,
		m.cached,
		m.content,
		m.created_at
    FROM medias m
        INNER JOIN entry_medias em ON m.id=em.media_id
        INNER JOIN entries e ON em.entry_id=e.id
        INNER JOIN feeds f ON e.feed_id=f.id
    WHERE 
        m.cached='T'
	GROUP BY m.id
	ORDER BY MIN(f.id), m.id ASC
	OFFSET $1
    LIMIT $2
`

	medias := make(map[string]model.Medias, 0)
	rows, err := s.db.Query(query, offset, limit)
	defer rows.Close()
	if err == sql.ErrNoRows {
		return medias, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to fetch cached medias: %v", err)
	}

	for rows.Next() {
		var feed string
		var media model.Media
		err := rows.Scan(
			&feed,
			&media.ID,
			&media.URL,
			&media.URLHash,
			&media.MimeType,
			&media.Cached,
			&media.Content,
			&media.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch cached medias row: %v", err)
		}
		medias[feed] = append(medias[feed], &media)
	}
	return medias, nil
}
