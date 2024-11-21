// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/locale"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/readability"
	"miniflux.app/v2/internal/reader/sanitizer"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) updateEntry(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryForm := form.NewEntryForm(r)

	feeds, err := h.store.Feeds(user.ID, false)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entryForm.Readability {
		_, content, err := readability.ExtractContent(strings.NewReader(entryForm.Content))
		if err == nil {
			entryForm.Content = content
		}
	}
	entryForm.Content = sanitizer.Sanitize(entryForm.URL, entryForm.Content)

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", entryForm)
	view.Set("feeds", feeds)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))

	if err := entryForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Translate(user.Language))
		html.OK(w, r, view.Render("edit_entry"))
		return
	}

	var entry *model.Entry

	// add entry
	if entryForm.EntryID == 0 {
		entry = entryForm.Merge(&model.Entry{
			UserID:  user.ID,
			Hash:    crypto.Hash(entryForm.URL),
			Date:    time.Now(),
			Status:  model.EntryStatusUnread,
			Starred: true,
		})

		if !h.store.EntryExists(entry) {
			tx, err := h.store.Begin()
			if err == nil {
				err = h.store.CreateEntry(tx, entry)
			}
			if err == nil {
				err = tx.Commit()
			}
			if err != nil {
				if tx != nil {
					tx.Rollback()
				}
				view.Set("errorMessage", err.Error())
				html.OK(w, r, view.Render("edit_entry"))
				return
			}
			_ = h.store.ToggleBookmark(entry.UserID, entry.ID)
		} else {
			builder := h.store.NewEntryQueryBuilder(user.ID)
			builder.WithEntryHash(entry.Hash)
			builder.WithFeedID(entry.FeedID)
			entry, err = builder.GetEntry()
			if err != nil {
				view.Set("errorMessage", err.Error())
				html.OK(w, r, view.Render("edit_entry"))
				return
			}
			view.Set("errorMessage", locale.NewLocalizedError("error.entry_existed"))
			view.Set("errorAction", route.Path(h.router, "editEntry", "entryID", entry.ID))
			html.OK(w, r, view.Render("edit_entry"))
			return
		}
		html.Redirect(w, r, route.Path(h.router, "readEntry", "entryID", entry.ID))
		return
	}
	// edit entry
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryForm.EntryID)
	entry, err = builder.GetEntry()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	entry.Starred = true
	entry.Status = model.EntryStatusUnread

	err = h.store.EditEntry(entryForm.Merge(entry))
	if err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_entry"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "readEntry", "entryID", entry.ID))
}
