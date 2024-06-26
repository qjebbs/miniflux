// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showStatPage(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nsfw := request.IsNSFWEnabled(r)

	builder := h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStatus(model.EntryStatusUnread)
	if nsfw {
		builder.WithoutNSFW()
	}
	countUnread, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	builder = h.store.NewEntryQueryBuilder(user.ID)
	builder.WithStarred(true)
	if nsfw {
		builder.WithoutNSFW()
	}
	countStarred, err := builder.CountEntries()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	var unreadByFeed, unreadByCategory, starredByFeed, starredByCategory, emptyStat model.EntryStat

	emptyStat = make(model.EntryStat, 0)
	if countUnread == 0 {
		unreadByFeed = emptyStat
		unreadByCategory = emptyStat
	} else {
		unreadByFeed, err = h.store.UnreadStatByFeed(user.ID, nsfw)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		unreadByCategory, err = h.store.UnreadStatByCategory(user.ID, nsfw)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}
	if countStarred == 0 {
		starredByFeed = emptyStat
		starredByCategory = emptyStat
	} else {
		starredByFeed, err = h.store.StarredStatByFeed(user.ID, nsfw)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}

		starredByCategory, err = h.store.StarredStatByCategory(user.ID, nsfw)
		if err != nil {
			html.ServerError(w, r, err)
			return
		}
	}

	view.Set("unreadByFeed", unreadByFeed)
	view.Set("unreadByCategory", unreadByCategory)
	view.Set("starredByFeed", starredByFeed)
	view.Set("starredByCategory", starredByCategory)
	view.Set("menu", "home")
	view.Set("user", user)
	view.Set("countUnread", countUnread)
	view.Set("countStarred", countStarred)
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("hasSaveEntry", h.store.HasSaveEntry(user.ID))

	html.OK(w, r, view.Render("stat"))
}
