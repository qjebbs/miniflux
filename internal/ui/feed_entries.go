// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showFeedEntriesPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.FeedByID(user.ID, feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	offset := request.QueryIntParam(r, "offset", 0)
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithFeedID(feed.ID)
	builder.WithStatus(model.EntryStatusUnread)
	builder.WithSorting(user.EntryOrder, user.EntryDirection)
	builder.WithSorting("id", user.EntryDirection)
	builder.WithOffset(offset)
	builder.WithLimit(user.EntriesPerPage)

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

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("feed", feed)
	view.Set("entries", entries)
	view.Set("total", count)
	view.Set("pagination", getPagination(route.Path(h.router, "feedEntries", "feedID", feed.ID), count, offset, user.EntriesPerPage))
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))
	view.Set("showOnlyUnreadEntries", true)
	view.Set("pageEntriesType", "unread")
	if feed.View != model.ViewDefault {
		view.Set("view", feed.View)
	}

	html.OK(w, r, view.Render("feed_entries"))
}
