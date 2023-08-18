// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/config"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/model"
	"miniflux.app/storage"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showReadEntryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryID)
	builder.WithoutStatus(model.EntryStatusRemoved)

	entry, err := builder.GetEntry()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	entryPaginationBuilder := storage.NewEntryPaginationBuilder(h.store, user.ID, entry.ID, "changed_at", "desc")
	entryPaginationBuilder.WithStatus(model.EntryStatusRead)
	prevEntry, nextEntry, err := entryPaginationBuilder.Entries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nextEntryRoute := ""
	if nextEntry != nil {
		nextEntryRoute = route.Path(h.router, "readEntry", "entryID", nextEntry.ID)
	}

	prevEntryRoute := ""
	if prevEntry != nil {
		prevEntryRoute = route.Path(h.router, "readEntry", "entryID", prevEntry.ID)
	}

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entry", entry)
	view.Set("prevEntry", prevEntry)
	view.Set("nextEntry", nextEntry)
	view.Set("nextEntryRoute", nextEntryRoute)
	view.Set("prevEntryRoute", prevEntryRoute)
	view.Set("menu", "history")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	if config.Opts.HasCacheService() {
		countMedias, countCached, _, err := h.store.MediaStatisticsByEntry(entryID)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
		view.Set("hasCacheService", countMedias > 0)
		view.Set("entryCached", countCached > 0)
	} else {
		view.Set("hasCacheService", false)
		view.Set("entryCached", false)
	}

	html.OK(w, r, view.Render("entry"))
}
