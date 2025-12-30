package storage // import "miniflux.app/v2/internal/storage"

import (
	"database/sql"
	"fmt"

	"miniflux.app/v2/internal/model"
)

// UnreadStatByFeed returns unread count of feeds.
func (s *Storage) UnreadStatByFeed(userID int64, nsfw bool) (stat model.EntryStat, err error) {
	query := `
		SELECT f.id, f.title, max(i.id) icon, max(i.external_id) icon_external_id, count(e.id) u_count
		FROM feeds f
			INNER JOIN categories c ON c.id=f.category_id
			LEFT JOIN (
				SELECT f.id id, count(e.id) count
				FROM feeds f
					INNER JOIN entries e ON f.id=e.feed_id
				WHERE f.user_id=$1 AND e.starred='T'
				GROUP BY f.id
			) starred ON f.id=starred.id
			INNER JOIN entries e ON f.id=e.feed_id
			LEFT JOIN feed_icons fi ON fi.feed_id=f.id
			LEFT JOIN icons i ON i.id=fi.icon_id
		WHERE f.user_id=$1 AND e.status='unread' %s
		GROUP BY f.id
		ORDER BY max(starred.count) DESC NULLS LAST, f.title ASC`
	nsfwCond := ""
	if nsfw {
		nsfwCond = "AND c.nsfw = 'f' AND f.nsfw = 'f'"
	}
	query = fmt.Sprintf(query, nsfwCond)
	return s.feedStatistics(query, userID)
}

// StarredStatByFeed returns starred count of feeds.
func (s *Storage) StarredStatByFeed(userID int64, nsfw bool) (stat model.EntryStat, err error) {
	query := `
		SELECT f.id, f.title, max(i.id) icon, max(i.external_id) icon_external_id, count(e.id) s_count
		FROM feeds f
			INNER JOIN entries e ON f.id=e.feed_id
			INNER JOIN categories c ON c.id=f.category_id
			LEFT JOIN feed_icons fi ON fi.feed_id=f.id
			LEFT JOIN icons i ON i.id=fi.icon_id
		WHERE f.user_id=$1 AND e.starred='T' %s
		GROUP BY f.id
		ORDER BY s_count DESC NULLS LAST, f.title ASC`
	nsfwCond := ""
	if nsfw {
		nsfwCond = "AND c.nsfw = 'f' AND f.nsfw = 'f'"
	}
	query = fmt.Sprintf(query, nsfwCond)
	return s.feedStatistics(query, userID)
}

// UnreadStatByCategory returns unread count of categories.
func (s *Storage) UnreadStatByCategory(userID int64, nsfw bool) (stat model.EntryStat, err error) {
	query := `
		SELECT c.id, c.title, count(e.id) u_count
		FROM categories c
			LEFT JOIN (
				SELECT c.id id, count(e.id) count
				FROM categories c
					INNER JOIN feeds f on c.id=f.category_id
					INNER JOIN entries e ON f.id=e.feed_id
				WHERE c.user_id=$1 AND e.starred='T'
				GROUP BY c.id
			) starred ON c.id=starred.id
			INNER JOIN feeds f on c.id=f.category_id
			INNER JOIN entries e ON f.id=e.feed_id
		WHERE c.user_id=$1 AND e.status='unread' %s
		GROUP BY c.id
		ORDER BY max(starred.count) DESC NULLS LAST, c.title ASC`
	nsfwCond := ""
	if nsfw {
		nsfwCond = "AND c.nsfw = 'f' AND f.nsfw = 'f'"
	}
	query = fmt.Sprintf(query, nsfwCond)
	return s.categoryStatistics(query, userID)
}

// StarredStatByCategory returns starred count of categories.
func (s *Storage) StarredStatByCategory(userID int64, nsfw bool) (stat model.EntryStat, err error) {
	query := `
		SELECT c.id, c.title, count(e.id) s_count
		FROM categories c
			INNER JOIN feeds f on c.id=f.category_id
			INNER JOIN entries e ON f.id=e.feed_id
		WHERE c.user_id=$1 AND e.starred='T' %s
		GROUP BY c.id
		ORDER BY s_count DESC NULLS LAST, c.title ASC`
	nsfwCond := ""
	if nsfw {
		nsfwCond = "AND c.nsfw = 'f' AND f.nsfw = 'f'"
	}
	query = fmt.Sprintf(query, nsfwCond)
	return s.categoryStatistics(query, userID)
}

func (s *Storage) feedStatistics(query string, args ...interface{}) (stat model.EntryStat, err error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get feed entry statistics: %v", err)
	}
	defer rows.Close()

	stat = make(model.EntryStat, 0)

	for rows.Next() {
		var iconID sql.NullInt64
		var externalIconID sql.NullString
		item := model.EntryStatItem{
			Feed: &model.Feed{
				Icon: &model.FeedIcon{},
			},
		}
		err := rows.Scan(
			&item.Feed.ID,
			&item.Feed.Title,
			&iconID,
			&externalIconID,
			&item.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to get feed statistics row: %v", err)
		}
		if iconID.Valid && externalIconID.Valid && externalIconID.String != "" {
			item.Feed.Icon.FeedID = item.Feed.ID
			item.Feed.Icon.IconID = iconID.Int64
			item.Feed.Icon.ExternalIconID = externalIconID.String
		} else {
			item.Feed.Icon.IconID = 0
		}
		stat = append(stat, &item)
	}
	return stat, nil
}

func (s *Storage) categoryStatistics(query string, args ...interface{}) (stat model.EntryStat, err error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to get category statistics: %v", err)
	}
	defer rows.Close()

	stat = make(model.EntryStat, 0)

	for rows.Next() {
		item := model.EntryStatItem{
			Category: &model.Category{},
		}
		err := rows.Scan(
			&item.Category.ID,
			&item.Category.Title,
			&item.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry statistics row: %v", err)
		}
		stat = append(stat, &item)
	}
	return stat, nil
}
