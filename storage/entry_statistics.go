package storage // import "miniflux.app/storage"

import (
	"fmt"
	"time"

	"miniflux.app/model"

	"miniflux.app/timer"
)

// UnreadStatByFeed returns unread count of feeds.
func (s *Storage) UnreadStatByFeed(userID int64) (stat model.EntryStat, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:UnreadStatByFeed]")
	return feedStatisticsByCond(s, userID, "e.status='unread'")
}

// StarredStatByFeed returns starred count of feeds.
func (s *Storage) StarredStatByFeed(userID int64) (stat model.EntryStat, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:StarredStatByFeed]")
	return feedStatisticsByCond(s, userID, "e.starred='T'")
}

// UnreadStatByCategory returns unread count of categories.
func (s *Storage) UnreadStatByCategory(userID int64) (stat model.EntryStat, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:UnreadStatByCategory]")
	return categoryStatisticsByCond(s, userID, "e.status='unread'")
}

// StarredStatByCategory returns starred count of categories.
func (s *Storage) StarredStatByCategory(userID int64) (stat model.EntryStat, err error) {
	defer timer.ExecutionTime(time.Now(), "[Storage:StarredStatByCategory]")
	return categoryStatisticsByCond(s, userID, "e.starred='T'")
}

func feedStatisticsByCond(store *Storage, userID int64, cond string) (stat model.EntryStat, err error) {
	query := fmt.Sprintf(`
	SELECT f.id, f.title, max(fi.icon_id) icon, count(e.id)
	FROM feeds f
		INNER JOIN entries e ON f.id=e.feed_id
		LEFT JOIN feed_icons fi ON fi.feed_id=f.id
	WHERE f.user_id=$1 AND %s
	GROUP BY f.id
	ORDER BY f.title ASC`, cond)

	rows, err := store.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get entry statistics: %v", err)
	}
	defer rows.Close()

	stat = make(model.EntryStat, 0)

	for rows.Next() {
		var iconID interface{}
		item := model.EntryStatItem{
			Feed: &model.Feed{
				Icon: &model.FeedIcon{},
			},
		}
		err := rows.Scan(
			&item.Feed.ID,
			&item.Feed.Title,
			&iconID,
			&item.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch entry statistics row: %v", err)
		}
		if iconID == nil {
			item.Feed.Icon.IconID = 0
		} else {
			item.Feed.Icon.IconID = iconID.(int64)
		}
		stat = append(stat, &item)
	}
	return stat, nil
}

func categoryStatisticsByCond(store *Storage, userID int64, cond string) (stat model.EntryStat, err error) {
	query := fmt.Sprintf(`
	SELECT c.id, c.title, count(e.id)
	FROM categories c
		INNER JOIN feeds f on c.id=f.category_id
		INNER JOIN entries e ON f.id=e.feed_id
	WHERE c.user_id=$1 AND %s
	GROUP BY c.id`, cond)

	rows, err := store.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to get entry statistics: %v", err)
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
