// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) showAddEntryPage(w http.ResponseWriter, r *http.Request) {
	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feeds, err := h.store.Feeds(user.ID, false)
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
	view.Set("feeds", feeds)
	view.Set("categories", categories)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("defaultUserAgent", client.DefaultUserAgent)
	view.Set("form", &form.EntryForm{FeedID: 0})

	html.OK(w, r, view.Render("add_entry"))
}
