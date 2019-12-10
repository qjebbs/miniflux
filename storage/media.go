package storage // import "miniflux.app/storage"

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/filesystem"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/reader/media"
	"miniflux.app/timer"
)

// MediaByURL returns an Media by the url.
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) MediaByURL(URL string) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.MediaByHash(m)
	return m, err
}

// MediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) MediaByHash(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:MediaByHash]")

	err := s.db.QueryRow(
		`SELECT id, url, mime_type, content, cached FROM medias WHERE url_hash=$1`,
		media.URLHash,
	).Scan(
		&media.ID,
		&media.URL,
		&media.MimeType,
		&media.Content,
		&media.Cached,
	)
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("Unable to fetch media by hash: %v", err)
	}

	return nil
}

// UserMediaByURL returns an Media by the url.
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) UserMediaByURL(URL string, userID int64) (*model.Media, error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByURL]")

	m := &model.Media{URLHash: media.URLHash(URL)}
	err := s.UserMediaByHash(m, userID)
	return m, err
}

// UserMediaByHash returns an Media by the url hash (checksum).
// Notice the media fetched could be an unsucessfully cached one.
// Remember to check Media.Cached.
func (s *Storage) UserMediaByHash(media *model.Media, userID int64) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UserMediaByHash]")

	useCache := false
	// "ORDER BY use_cache DESC" is important.
	// It makes sure the result of use_cache in this query would be true
	// if any record set "use_cache" to true.
	// e.g.: One image could be used in multiple entries to a single user,
	// if one of any records uses cache, then rest of them use cache as well
	err := s.db.QueryRow(`
	SELECT m.id, m.url, m.mime_type, m.content, m.cached, e.url, em.use_cache
	FROM medias m
		INNER JOIN entry_medias em ON m.id=em.media_id
		INNER JOIN entries e ON e.id=em.entry_id
		INNER JOIN feeds f on f.id=e.feed_id
	WHERE m.url_hash=$1 AND f.user_id=$2
	ORDER BY use_cache DESC
`,
		media.URLHash,
		userID,
	).Scan(
		&media.ID,
		&media.URL,
		&media.MimeType,
		&media.Content,
		&media.Cached,
		&media.Referrer,
		&useCache,
	)
	// If not useCache, the the media cache could be created by other users.
	media.Cached = media.Cached && useCache
	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return fmt.Errorf("Unable to fetch media by hash: %v", err)
	}

	return nil
}

// CreateMedia creates a new media item.
func (s *Storage) CreateMedia(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:CreateMedia]")
	query := `
	INSERT INTO medias
	(url, url_hash, mime_type, content, size, cached)
	VALUES
	($1, $2, $3, $4, $5, $6)
	RETURNING id
`
	err := s.db.QueryRow(
		query,
		media.URL,
		media.URLHash,
		normalizeMimeType(media.MimeType),
		media.Content,
		media.Size,
		media.Cached,
	).Scan(&media.ID)

	if err != nil {
		return fmt.Errorf("Unable to create media: %v", err)
	}

	return nil
}

// CreateEntryMedia creates media for a single entry, and it won't replace existed media or media references
func (s *Storage) CreateEntryMedia(entry *model.Entry) ([]int64, error) {
	cap := 15
	medias := make(map[string]string, cap)
	var mediaIDs []int64

	urls, err := media.ParseDocument(entry)
	if err != nil || len(urls) == 0 {
		return nil, err
	}
	for _, u := range urls {
		hash := media.URLHash(u)
		if _, ok := medias[hash]; !ok {
			medias[hash] = fmt.Sprintf("('%v','%v'),", strings.Replace(u, "'", "''", -1), hash)
		}
	}

	if len(medias) == 0 {
		return nil, nil
	}
	// insert medias records
	var buf bytes.Buffer
	for _, em := range medias {
		buf.WriteString(em)
	}
	vals := buf.String()[:buf.Len()-1]
	sql := fmt.Sprintf(`
		INSERT INTO medias (url, url_hash)
		VALUES %s
		ON CONFLICT (url_hash) DO UPDATE
			SET created_at=current_timestamp
		RETURNING id, url_hash
	`, vals)
	rows, err := s.db.Query(sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var m model.Media
		err = rows.Scan(&m.ID, &m.URLHash)
		if err != nil {
			return nil, err
		}
		mediaIDs = append(mediaIDs, m.ID)
	}

	// insert entry_medias records
	buf.Reset()
	for _, id := range mediaIDs {
		buf.WriteString(fmt.Sprintf("(%v,%v),", entry.ID, id))
	}
	vals = buf.String()[:buf.Len()-1]
	sql = fmt.Sprintf(`
		INSERT INTO entry_medias (entry_id, media_id) 
		VALUES %s
		ON CONFLICT DO NOTHING`, vals)
	_, err = s.db.Exec(sql)
	return mediaIDs, err
}

// CreateEntriesMedia creates media for a slice of entries at a time
func (s *Storage) CreateEntriesMedia(entries model.Entries) error {
	var err error
	cap := len(entries) * 15
	medias := make(map[string]string, cap)
	type IDSet map[int64]*int8
	// one url could be in multiple entries, and even appears many times in one entry
	// use IDSET to make sure one url could have multiple entries, but not duplicated
	urlEntries := make(map[string]IDSet, cap)
	mediaIDs := make(map[string]int64, cap)
	for _, entry := range entries {
		urls, err := media.ParseDocument(entry)
		if err != nil || len(urls) == 0 {
			continue
		}
		for _, u := range urls {
			hash := media.URLHash(u)
			if _, ok := medias[hash]; !ok {
				medias[hash] = fmt.Sprintf("('%v','%v'),", strings.Replace(u, "'", "''", -1), hash)
			}
			if _, ok := urlEntries[hash]; !ok {
				urlEntries[hash] = make(IDSet, 0)
			}
			urlEntries[hash][entry.ID] = nil
		}
	}

	if len(medias) == 0 {
		return nil
	}
	// insert medias records
	var buf bytes.Buffer
	for _, em := range medias {
		buf.WriteString(em)
	}
	vals := buf.String()[:buf.Len()-1]
	sql := fmt.Sprintf(`
		INSERT INTO medias (url, url_hash)
		VALUES %s
		ON CONFLICT (url_hash) DO UPDATE
			SET created_at=current_timestamp
		RETURNING id, url_hash
	`, vals)
	rows, err := s.db.Query(sql)
	defer rows.Close()
	if err != nil {
		return err
	}
	for rows.Next() {
		var m model.Media
		err = rows.Scan(&m.ID, &m.URLHash)
		if err != nil {
			return err
		}
		mediaIDs[m.URLHash] = m.ID
	}

	// insert entry_medias records
	buf.Reset()
	for hash, idSet := range urlEntries {
		for id := range idSet {
			buf.WriteString(fmt.Sprintf("(%v,%v),", id, mediaIDs[hash]))
		}
	}
	vals = buf.String()[:buf.Len()-1]
	sql = fmt.Sprintf(`
		INSERT INTO entry_medias (entry_id, media_id) 
		VALUES %s
		ON CONFLICT DO NOTHING`, vals)
	_, err = s.db.Exec(sql)
	return err
}

// UpdateMedia updates a media cache.
func (s *Storage) UpdateMedia(media *model.Media) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateMedia]")
	query := `
	UPDATE medias
	SET mime_type=$2, content=$3, size=$4, cached=$5
	WHERE id = $1
`

	var err error
	if config.Opts.CacheLocation() != "database" {
		err = filesystem.SaveMediaFile(media)
		if err != nil {
			return fmt.Errorf("Unable to update media: %v", err)
		}
		_, err = s.db.Exec(
			query,
			media.ID,
			normalizeMimeType(media.MimeType),
			nil,
			media.Size,
			media.Cached,
		)
	} else {
		_, err = s.db.Exec(
			query,
			media.ID,
			normalizeMimeType(media.MimeType),
			media.Content,
			media.Size,
			media.Cached,
		)
	}

	if err != nil {
		return fmt.Errorf("Unable to update media: %v", err)
	}

	return nil
}

// UpdateMediaError updates media errors.
func (s *Storage) UpdateMediaError(m *model.Media) (err error) {
	query := `
		UPDATE
			medias
		SET
			error_count=$1
		WHERE
			id=$2
	`
	_, err = s.db.Exec(query,
		m.ErrorCount,
		m.ID,
	)

	if err != nil {
		return fmt.Errorf(`store: unable to update media error #%d: %v`, m.ID, err)
	}

	return nil
}

// Medias returns all media caches tht belongs to a user.
func (s *Storage) Medias(userID int64) (model.Medias, error) {
	defer timer.ExecutionTime(time.Now(), fmt.Sprintf("[Storage:Medias] userID=%d", userID))
	query := `
		SELECT
		m.id, m.url_hash, m.mime_type, m.content
		FROM medias m
		LEFT JOIN entry_medias em ON em.media_id=medias.id
		LEFT JOIN entries e ON e.id=em.entry_id
		WHERE m.cached='t' AND e.user_id=$1
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch medias: %v", err)
	}
	defer rows.Close()

	var medias model.Medias
	for rows.Next() {
		var media model.Media
		err := rows.Scan(&media.ID, &media.URLHash, &media.MimeType, &media.Content)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch medias row: %v", err)
		}
		medias = append(medias, &media)
	}

	return medias, nil
}

// UpdateEntryMedia updates media records for given entries
func (s *Storage) UpdateEntryMedia(entry *model.Entry) error {
	defer timer.ExecutionTime(time.Now(), "[Storage:UpdateEntryMedias]")
	if entry.Status == "" {
		err := s.db.QueryRow(`SELECT status FROM entries WHERE id=$1`, entry.ID).Scan(&entry.Status)
		if err != nil {
			return fmt.Errorf("[Storage:UpdateEntryMedias] unable to fetch entry status #%d: %v", entry.ID, err)
		}
	}
	if entry.Status == model.EntryStatusRemoved {
		return nil
	}
	mediaIDs, err := s.CreateEntryMedia(entry)
	if err != nil {
		return fmt.Errorf("[Storage:UpdateEntryMedias] unable to create media for entry #%d: %v", entry.ID, err)
	}
	if len(mediaIDs) == 0 {
		// no media for update entry, remove all its references
		_, err = s.db.Exec(`DELETE FROM entry_medias WHERE entry_id = $1`, entry.ID)
	} else {
		// remove references of removed media
		var buf bytes.Buffer
		for _, id := range mediaIDs {
			buf.WriteString(fmt.Sprintf("%d,", id))
		}
		vals := buf.String()[:buf.Len()-1]
		query := fmt.Sprintf(`DELETE FROM entry_medias WHERE entry_id = %d AND media_id not in (%s)`, entry.ID, vals)
		_, err = s.db.Exec(query)
	}
	if err != nil {
		return fmt.Errorf("unable to clean media for entry #%d: %v", entry.ID, err)
	}
	return nil
}

// cleanMediaReferences removes entry_medias records which belong to entries that are marked as removed.
func (s *Storage) cleanMediaReferences() error {
	query := `
		DELETE FROM entry_medias
		WHERE entry_id IN (
			SELECT id FROM entries WHERE status=$1
		)
	`
	result, err := s.db.Exec(query, model.EntryStatusRemoved)
	if err != nil {
		return fmt.Errorf("store: unable to clean media references: %v", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(`store:unable to clean media references: %v`, err)
	}
	logger.Info("%d unused media references removed.", count)

	return nil
}

// cleanMediaRecords removes media records which has no reference record at all.
// Important: this should always run after CleanMediaCaches(), or caches in disk will be orphan files.
func (s *Storage) cleanMediaRecords() error {
	query := `
		DELETE FROM medias
		WHERE id IN (
			SELECT m.id 
			FROM medias m
			LEFT JOIN entry_medias em on m.id=em.media_id
			WHERE em.entry_id IS NULL
		)
	`
	result, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("unable to clean media records: %v", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to clean media records: %v", err)
	}

	logger.Info("%d unused media records removed.", count)
	return nil
}
