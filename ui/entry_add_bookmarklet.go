// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) bookmarkletEntry(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	feeds, err := h.store.Feeds(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	bookmarkletURL := request.QueryStringParam(r, "uri", "")

	view.Set("feeds", feeds)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, request.IsNSFWEnabled(r)))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	view.Set("defaultUserAgent", client.DefaultUserAgent)
	view.Set("form", &form.EntryForm{URL: bookmarkletURL})

	html.OK(w, r, view.Render("add_entry"))
}
