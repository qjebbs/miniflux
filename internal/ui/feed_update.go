// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
	"miniflux.app/v2/internal/validator"
)

func (h *handler) updateFeed(w http.ResponseWriter, r *http.Request) {
	loggedUser, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.FeedByID(loggedUser.ID, feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	categories, err := h.store.Categories(loggedUser.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedForm := form.NewFeedForm(r)

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", loggedUser)
	view.Set("countUnread", h.store.CountUnreadEntries(loggedUser.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(loggedUser.ID, nsfw))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())

	feedModificationRequest := &model.FeedModificationRequest{
		FeedURL:         model.OptionalString(feedForm.FeedURL),
		SiteURL:         model.OptionalString(feedForm.SiteURL),
		Title:           model.OptionalString(feedForm.Title),
		CategoryID:      model.OptionalNumber(feedForm.CategoryID),
		BlocklistRules:  model.OptionalString(feedForm.BlocklistRules),
		KeeplistRules:   model.OptionalString(feedForm.KeeplistRules),
		UrlRewriteRules: model.OptionalString(feedForm.UrlRewriteRules),
	}

	if validationErr := validator.ValidateFeedModification(h.store, loggedUser.ID, feed.ID, feedModificationRequest); validationErr != nil {
		view.Set("errorMessage", validationErr.Translate(loggedUser.Language))
		html.OK(w, r, view.Render("edit_feed"))
		return
	}

	err = h.store.UpdateFeed(feedForm.Merge(feed))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
}
