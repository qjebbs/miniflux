// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"
	"strings"

	"miniflux.app/reader/readability"

	"miniflux.app/http/request"
	"miniflux.app/http/response/html"
	"miniflux.app/http/route"
	"miniflux.app/logger"
	"miniflux.app/model"
	"miniflux.app/ui/form"
	"miniflux.app/ui/session"
	"miniflux.app/ui/view"
)

func (h *handler) updateEntry(w http.ResponseWriter, r *http.Request) {
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

	feeds, err := h.store.Feeds(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	entryForm := form.NewEntryForm(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", entryForm)
	view.Set("entry", entry)
	view.Set("feeds", feeds)
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID))
	view.Set("countErrorFeeds", h.store.CountErrorFeeds(user.ID))

	if err := entryForm.ValidateModification(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_entry"))
		return
	}
	if entryForm.Readability {
		content, err := readability.ExtractContent(strings.NewReader(entryForm.Content))
		if err == nil {
			entryForm.Content = content
		}
	}
	err = h.store.UpdateEntryByID(entryForm.Merge(entry))
	if err != nil {
		logger.Error("[UI:UpdateEntry] %v", err)
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("edit_entry"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "readEntry", "entryID", entry.ID))
}
