// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/reader/scraper"

	"miniflux.app/http/client"
	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) submitEntry(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	v := view.New(h.tpl, r, sess)

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

	v.Set("feeds", feeds)
	v.Set("user", user)
	v.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	v.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))
	v.Set("defaultUserAgent", client.DefaultUserAgent)

	entryForm := form.NewEntryForm(r)
	if err := entryForm.Validate(); err != nil {
		v.Set("form", entryForm)
		v.Set("errorMessage", err.Error())
		html.OK(w, r, v.Render("add_entry"))
		return
	}

	entry, err := scraper.FetchEntry(entryForm.URL, "", entryForm.UserAgent)
	entry.UserID = user.ID
	entry.FeedID = entryForm.FeedID

	if err != nil {
		logger.Error("[UI:ProcessEntryWebPage] %s", err)
		v.Set("form", entryForm)
		v.Set("errorMessage", err)
		html.OK(w, r, v.Render("add_entry"))
		return
	}
	if !h.store.EntryExists(entry) {
		err = h.store.CreateEntry(entry)
		if err != nil {
			v.Set("form", entryForm)
			v.Set("errorMessage", err)
			html.OK(w, r, v.Render("add_entry"))
			return
		}
	} else {
		builder := h.store.NewEntryQueryBuilder(user.ID)
		builder.WithEntryHash(entry.Hash)
		builder.WithFeedID(entry.FeedID)
		entry, err = builder.GetEntry()
		if err != nil {
			if err != nil {
				v.Set("form", entryForm)
				v.Set("errorMessage", err)
				html.OK(w, r, v.Render("add_entry"))
				return
			}
		}
	}
	html.Redirect(w, r, route.Path(h.router, "editEntry", "entryID", entry.ID))
}
