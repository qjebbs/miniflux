// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/model"
)

// FeedForm represents a feed form in the UI
type FeedForm struct {
	FeedURL         string
	SiteURL         string
	Title           string
	ScraperRules    string
	RewriteRules    string
	BlocklistRules  string
	KeeplistRules   string
	Crawler         bool
	CacheMedia      bool
	UserAgent       string
	CategoryID      int64
	Username        string
	Password        string
	View            string
	IgnoreHTTPCache bool
	FetchViaProxy   bool
	Disabled        bool
	NSFW            bool
}

// Merge updates the fields of the given feed.
func (f FeedForm) Merge(feed *model.Feed) *model.Feed {
	feed.Category.ID = f.CategoryID
	feed.Title = f.Title
	feed.SiteURL = f.SiteURL
	feed.FeedURL = f.FeedURL
	feed.ScraperRules = f.ScraperRules
	feed.RewriteRules = f.RewriteRules
	feed.BlocklistRules = f.BlocklistRules
	feed.KeeplistRules = f.KeeplistRules
	feed.Crawler = f.Crawler
	feed.CacheMedia = f.CacheMedia
	feed.UserAgent = f.UserAgent
	feed.ParsingErrorCount = 0
	feed.ParsingErrorMsg = ""
	feed.Username = f.Username
	feed.Password = f.Password
	feed.View = f.View
	feed.IgnoreHTTPCache = f.IgnoreHTTPCache
	feed.FetchViaProxy = f.FetchViaProxy
	feed.Disabled = f.Disabled
	feed.NSFW = f.NSFW
	return feed
}

// NewFeedForm parses the HTTP request and returns a FeedForm
func NewFeedForm(r *http.Request) *FeedForm {
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil {
		categoryID = 0
	}
	view := r.FormValue("view")
	if _, ok := model.Views()[view]; !ok {
		view = model.ViewDefault
	}
	return &FeedForm{
		FeedURL:         r.FormValue("feed_url"),
		SiteURL:         r.FormValue("site_url"),
		Title:           r.FormValue("title"),
		ScraperRules:    r.FormValue("scraper_rules"),
		UserAgent:       r.FormValue("user_agent"),
		RewriteRules:    r.FormValue("rewrite_rules"),
		BlocklistRules:  r.FormValue("blocklist_rules"),
		KeeplistRules:   r.FormValue("keeplist_rules"),
		Crawler:         r.FormValue("crawler") == "1",
		CacheMedia:      r.FormValue("cache_media") == "1",
		CategoryID:      int64(categoryID),
		Username:        r.FormValue("feed_username"),
		Password:        r.FormValue("feed_password"),
		View:            view,
		IgnoreHTTPCache: r.FormValue("ignore_http_cache") == "1",
		FetchViaProxy:   r.FormValue("fetch_via_proxy") == "1",
		Disabled:        r.FormValue("disabled") == "1",
		NSFW:            r.FormValue("nsfw") == "1",
	}
}
