package cli

import (
	"miniflux.app/v2/logger"
	"miniflux.app/v2/model"
	"miniflux.app/v2/storage"
)

func fixCovers(store *storage.Storage, uid int64) error {
	q := store.NewEntryQueryBuilder(uid)
	q.WithStatuses([]string{model.EntryStatusUnread})
	if err := fixEntries(store, q); err != nil {
		return err
	}
	q = store.NewEntryQueryBuilder(uid)
	q.WithoutStatus(model.EntryStatusUnread)
	q.WithStarred(true)
	if err := fixEntries(store, q); err != nil {
		return err
	}
	return nil
}

func fixEntries(store *storage.Storage, q *storage.EntryQueryBuilder) error {
	count, err := q.CountEntries()
	if err != nil {
		return err
	}
	offset := 0
	for offset < count {
		entries, err := q.WithOffset(offset).WithLimit(100).GetEntries()
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.CoverImage != "" {
				continue
			}
			if err := store.UpdateEntryContent(entry); err != nil {
				return err
			}
		}
		offset += 100
		logger.Info("fixed %d entries out of %d", offset, count)
	}
	logger.Info("all %d covers fixed", count)
	return nil
}
