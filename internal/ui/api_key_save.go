// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) saveAPIKey(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	apiKeyForm := form.NewAPIKeyForm(r)
	apiKeyCreationRequest := &model.APIKeyCreationRequest{
		Description: apiKeyForm.Description,
	}
	nsfw := request.IsNSFWEnabled(r)

	if validationErr := validator.ValidateAPIKeyCreation(h.store, user.ID, apiKeyCreationRequest); validationErr != nil {
		sess := session.New(h.store, request.SessionID(r))
		view := view.New(h.tpl, r, sess)
		view.Set("form", apiKeyForm)
		view.Set("menu", "settings")
		view.Set("user", user)
		view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
		view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
		view.Set("errorMessage", validationErr.Translate(user.Language))
		html.OK(w, r, view.Render("create_api_key"))
		return
	}

	if _, err = h.store.CreateAPIKey(user.ID, apiKeyCreationRequest.Description); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "apiKeys"))
}
