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

func (h *handler) saveCategory(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categoryForm := form.NewCategoryForm(r)

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", categoryForm)
	view.Set("menu", "categories")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID, nsfw))

	categoryCreationRequest := &model.CategoryCreationRequest{
		Title: categoryForm.Title,
		NSFW:  categoryForm.NSFW,
		View:  categoryForm.View,
	}

	if validationErr := validator.ValidateCategoryCreation(h.store, loggedUser.ID, categoryCreationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("create_category"))
		return
	}

	if _, err = h.store.CreateCategory(loggedUser.ID, categoryCreationRequest); err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "categories"))
}
