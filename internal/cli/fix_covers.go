package cli

import (
	"log/slog"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/storage"
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
			if err := store.UpdateEntryTitleAndContent(entry); err != nil {
				return err
			}
		}
		offset += 100
		slog.Info(
			"fix entries",
			slog.Int("fixed", offset), slog.Int("all", count),
		)
	}
	slog.Info("all covers fixed")
	return nil
}
