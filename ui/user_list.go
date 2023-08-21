// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"net/http"

	"miniflux.app/v2/http/request"
	"miniflux.app/v2/http/response/html"
	"miniflux.app/v2/ui/session"
	"miniflux.app/v2/ui/view"
)

func (h *handler) showUsersPage(w http.ResponseWriter, r *http.Request) {
	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if !user.IsAdmin {
		html.Forbidden(w, r)
		return
	}

	users, err := h.store.Users()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	users.UseTimezone(user.Timezone)

	view.Set("users", users)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))

	html.OK(w, r, view.Render("users"))
}
