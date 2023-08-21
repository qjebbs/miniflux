// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"net/http"

	"miniflux.app/v2/http/request"
	"miniflux.app/v2/http/response/html"
	"miniflux.app/v2/http/route"
	"miniflux.app/v2/logger"
	"miniflux.app/v2/model"
	"miniflux.app/v2/ui/form"
	"miniflux.app/v2/ui/session"
	"miniflux.app/v2/ui/view"
)

func (h *handler) saveAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	apiKeyForm := form.NewAPIKeyForm(r)
	nsfw := request.IsNSFWEnabled(r)

	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", apiKeyForm)
	view.Set("menu", "settings")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))

	if err := apiKeyForm.Validate(); err != nil {
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	if h.store.APIKeyExists(user.ID, apiKeyForm.Description) {
		view.Set("errorMessage", "error.api_key_already_exists")
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	apiKey := model.NewAPIKey(user.ID, apiKeyForm.Description)
	if err = h.store.CreateAPIKey(apiKey); err != nil {
		logger.Error("[UI:SaveAPIKey] %v", err)
		view.Set("errorMessage", "error.unable_to_create_api_key")
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "apiKeys"))
}
