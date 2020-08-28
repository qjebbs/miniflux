// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showEditEntryPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryID := request.RouteInt64Param(r, "entryID")
	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithEntryID(entryID)
	entry, err := builder.GetEntry()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if entry == nil {
		html.NotFound(w, r)
		return
	}

	var feeds model.Feeds
	nsfw := request.IsNSFWEnabled(r)
	feeds, err = h.store.Feeds(user.ID, false)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	categories := make(model.Categories, 0)
	encountered := make(map[int64]struct{})
	for _, feed := range feeds {
		if _, ok := encountered[feed.Category.ID]; !ok {
			categories = append(categories, feed.Category)
			encountered[feed.Category.ID] = struct{}{}
		}
	}

	entryForm := form.EntryForm{
		EntryID:     entry.ID,
		FeedID:      entry.FeedID,
		Title:       entry.Title,
		URL:         entry.URL,
		CommentsURL: entry.CommentsURL,
		Content:     entry.Content,
		Author:      entry.Author,
	}

	view.Set("form", entryForm)
	view.Set("feeds", feeds)
	view.Set("categories", categories)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID, nsfw))

	html.OK(w, r, view.Render("edit_entry"))
}
