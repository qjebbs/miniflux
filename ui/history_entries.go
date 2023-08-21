// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"net/http"

	"miniflux.app/v2/http/request"
	"miniflux.app/v2/http/response/html"
	"miniflux.app/v2/http/route"
	"miniflux.app/v2/model"
	"miniflux.app/v2/ui/session"
	"miniflux.app/v2/ui/view"
)

func (h *handler) showHistoryPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nsfw := request.IsNSFWEnabled(r)

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusRead)
	builder.WithSorting("changed_at", "DESC")
	builder.WithSorting("published_at", "DESC")
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)
	if nsfw {
		builder.WithoutNSFW()
	}

	entries, err := builder.GetEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	count, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("entries", entries)
	view.Set("total", count)
	view.Set("pagination", getPagination(route.Path(h.router, "history"), count, offset, user.EntriesPerPage))
	view.Set("menu", "history")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("pageEntriesType", "all")

	html.OK(w, r, view.Render("history_entries"))
}
