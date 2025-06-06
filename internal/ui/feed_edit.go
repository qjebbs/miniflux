// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"fmt"
	"net/http"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/ui/form"
	"miniflux.app/v2/internal/ui/session"
	"miniflux.app/v2/internal/ui/view"
)

func (h *handler) showEditFeedPage(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedID := request.RouteInt64Param(r, "feedID")
	feed, err := h.store.FeedByID(user.ID, feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	if feed == nil {
		html.NotFound(w, r)
		return
	}

	categories, err := h.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	feedForm := form.FeedForm{
		SiteURL:                     feed.SiteURL,
		FeedURL:                     feed.FeedURL,
		Title:                       feed.Title,
		Description:                 feed.Description,
		ScraperRules:                feed.ScraperRules,
		RewriteRules:                feed.RewriteRules,
		BlocklistRules:              feed.BlocklistRules,
		KeeplistRules:               feed.KeeplistRules,
		UrlRewriteRules:             feed.UrlRewriteRules,
		Crawler:                     feed.Crawler,
		CacheMedia:                  feed.CacheMedia,
		UserAgent:                   feed.UserAgent,
		Cookie:                      feed.Cookie,
		CategoryID:                  feed.Category.ID,
		Username:                    feed.Username,
		Password:                    feed.Password,
		View:                        feed.View,
		IgnoreHTTPCache:             feed.IgnoreHTTPCache,
		AllowSelfSignedCertificates: feed.AllowSelfSignedCertificates,
		FetchViaProxy:               feed.FetchViaProxy,
		Disabled:                    feed.Disabled,
		NoMediaPlayer:               feed.NoMediaPlayer,
		NSFW:                        feed.NSFW,
		ProxifyMedia:                feed.ProxifyMedia,
		AppriseServiceURLs:          feed.AppriseServiceURLs,
		WebhookURL:                  feed.WebhookURL,
		DisableHTTP2:                feed.DisableHTTP2,
		NtfyEnabled:                 feed.NtfyEnabled,
		NtfyPriority:                feed.NtfyPriority,
		PushoverEnabled:             feed.PushoverEnabled,
		PushoverPriority:            feed.PushoverPriority,
	}

	all, count, size, err := h.store.MediaStatisticsByFeed(feedID)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("form", feedForm)
	view.Set("categories", categories)
	view.Set("views", model.Views())
	view.Set("feed", feed)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))
	view.Set("defaultUserAgent", config.Opts.HTTPClientUserAgent())
	view.Set("mediaCount", all)
	view.Set("cacheCount", count)
	view.Set("cacheSize", byteSizeHumanReadable(size))
	view.Set("hasProxyConfigured", config.Opts.HasHTTPClientProxyConfigured())

	html.OK(w, r, view.Render("edit_feed"))
}

func byteSizeHumanReadable(size int) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
		PB
		EB
		ZB
		YB
	)
	unit := ""
	sz := float64(size)
	if sz < KB {
		unit = "B"
	} else if sz < MB {
		unit = "KB"
		sz = sz / KB
	} else if sz < GB {
		unit = "MB"
		sz = sz / MB
	} else if sz < TB {
		unit = "GB"
		sz = sz / GB
	} else if sz < PB {
		unit = "TB"
		sz = sz / TB
	} else if sz < EB {
		unit = "PB"
		sz = sz / PB
	} else if sz < ZB {
		unit = "EB"
		sz = sz / EB
	} else if sz < YB {
		unit = "ZB"
		sz = sz / ZB
	} else {
		unit = "YB"
		sz = sz / YB
	}
	return fmt.Sprintf("%.2f%s", sz, unit)
}
