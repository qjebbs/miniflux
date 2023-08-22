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
	feedHandler "miniflux.app/v2/internal/reader/handler"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showChooseSubscriptionPage(w http.ResponseWriter, r *http.Request) {
	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)

	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	categories, err := h.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	view.Set("categories", categories)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())

	subscriptionForm := form.NewSubscriptionForm(r)
	if err := subscriptionForm.Validate(); err != nil {
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", err.Error())
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	feed, err := feedHandler.CreateFeed(h.store, user.ID, &model.FeedCreationRequest{
		CategoryID:                  subscriptionForm.CategoryID,
		FeedURL:                     subscriptionForm.URL,
		Crawler:                     subscriptionForm.Crawler,
		AllowSelfSignedCertificates: subscriptionForm.AllowSelfSignedCertificates,
		UserAgent:                   subscriptionForm.UserAgent,
		Cookie:                      subscriptionForm.Cookie,
		Username:                    subscriptionForm.Username,
		Password:                    subscriptionForm.Password,
		ScraperRules:                subscriptionForm.ScraperRules,
		RewriteRules:                subscriptionForm.RewriteRules,
		BlocklistRules:              subscriptionForm.BlocklistRules,
		KeeplistRules:               subscriptionForm.KeeplistRules,
		UrlRewriteRules:             subscriptionForm.UrlRewriteRules,
		FetchViaProxy:               subscriptionForm.FetchViaProxy,
		NSFW:                        subscriptionForm.NSFW,
	})
	if err != nil {
		view.Set("form", subscriptionForm)
		view.Set("errorMessage", err)
		html.OK(w, r, view.Render("add_subscription"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feedEntries", "feedID", feed.ID))
}
