// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/scraper"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/logger"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) submitEntry(w http.ResponseWriter, r *http.Request) {
	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	v := view.New(h.tpl, r, sess)

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

	entryForm := form.NewEntryForm(r)

	v.Set("form", entryForm)
	v.Set("feeds", feeds)
	v.Set("categories", categories)
	v.Set("user", user)
	v.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	v.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	v.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())

	if err := entryForm.Validate(); err != nil {
		v.Set("errorMessage", err.Error())
		html.OK(w, r, v.Render("add_entry"))
		return
	}

	entry, err := scraper.FetchEntry(entryForm.URL, "", entryForm.UserAgent, entryForm.Cookies)
	if err != nil {
		logger.Error("[UI:ProcessEntryWebPage] %s", err)
		v.Set("errorMessage", err)
		html.OK(w, r, v.Render("add_entry"))
		return
	}

	entryForm.Title = entry.Title
	entryForm.Content = entry.Content
	html.OK(w, r, v.Render("edit_entry"))
}
