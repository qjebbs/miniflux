// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

import (
	"fmt"
	"io"
	"math"
	"time"

	"miniflux.app/v2/internal/config"
)

// List of supported schedulers.
const (
	SchedulerRoundRobin     = "round_robin"
	SchedulerEntryFrequency = "entry_frequency"
	// Default settings for the feed query builder
	DefaultFeedSorting          = "parsing_error_count"
	DefaultFeedSortingDirection = "desc"
)

// Feed represents a feed in the application.
type Feed struct {
	ID                          int64     `json:"id"`
	UserID                      int64     `json:"user_id"`
	FeedURL                     string    `json:"feed_url"`
	SiteURL                     string    `json:"site_url"`
	Title                       string    `json:"title"`
	CheckedAt                   time.Time `json:"checked_at"`
	NextCheckAt                 time.Time `json:"next_check_at"`
	EtagHeader                  string    `json:"etag_header"`
	LastModifiedHeader          string    `json:"last_modified_header"`
	ParsingErrorMsg             string    `json:"parsing_error_message"`
	ParsingErrorCount           int       `json:"parsing_error_count"`
	ScraperRules                string    `json:"scraper_rules"`
	RewriteRules                string    `json:"rewrite_rules"`
	Crawler                     bool      `json:"crawler"`
	BlocklistRules              string    `json:"blocklist_rules"`
	KeeplistRules               string    `json:"keeplist_rules"`
	UrlRewriteRules             string    `json:"urlrewrite_rules"`
	UserAgent                   string    `json:"user_agent"`
	Cookie                      string    `json:"cookie"`
	Username                    string    `json:"username"`
	Password                    string    `json:"password"`
	Disabled                    bool      `json:"disabled"`
	NoMediaPlayer               bool      `json:"no_media_player"`
	IgnoreHTTPCache             bool      `json:"ignore_http_cache"`
	AllowSelfSignedCertificates bool      `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool      `json:"fetch_via_proxy"`
	NSFW                        bool      `json:"nsfw"`
	AppriseServiceURLs          string    `json:"apprise_service_urls"`
	DisableHTTP2                bool      `json:"disable_http2"`

	// Non persisted attributes
	Category *Category `json:"category,omitempty"`
	Icon     *FeedIcon `json:"icon"`
	Entries  Entries   `json:"entries,omitempty"`

	TTL                    int    `json:"-"`
	IconURL                string `json:"-"`
	UnreadCount            int    `json:"-"`
	ReadCount              int    `json:"-"`
	NumberOfVisibleEntries int    `json:"-"`

	View         string `json:"view"`
	CacheMedia   bool   `json:"cache_media"`
	ProxifyMedia bool   `json:"-"`
}

type FeedCounters struct {
	ReadCounters   map[int64]int `json:"reads"`
	UnreadCounters map[int64]int `json:"unreads"`
}

func (f *Feed) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, FeedURL=%s, SiteURL=%s, Title=%s, Category={%s}",
		f.ID,
		f.UserID,
		f.FeedURL,
		f.SiteURL,
		f.Title,
		f.Category,
	)
}

// WithCategoryID initializes the category attribute of the feed.
func (f *Feed) WithCategoryID(categoryID int64) {
	f.Category = &Category{ID: categoryID}
}

// WithTranslatedErrorMessage adds a new error message and increment the error counter.
func (f *Feed) WithTranslatedErrorMessage(message string) {
	f.ParsingErrorCount++
	f.ParsingErrorMsg = message
}

// ResetErrorCounter removes all previous errors.
func (f *Feed) ResetErrorCounter() {
	f.ParsingErrorCount = 0
	f.ParsingErrorMsg = ""
}

// CheckedNow set attribute values when the feed is refreshed.
func (f *Feed) CheckedNow() {
	f.CheckedAt = time.Now()

	if f.SiteURL == "" {
		f.SiteURL = f.FeedURL
	}
}

// ScheduleNextCheck set "next_check_at" of a feed based on the scheduler selected from the configuration.
func (f *Feed) ScheduleNextCheck(weeklyCount int, newTTL int) {
	f.TTL = newTTL
	// Default to the global config Polling Frequency.
	var intervalMinutes int
	switch config.Opts.PollingScheduler() {
	case SchedulerEntryFrequency:
		if weeklyCount <= 0 {
			intervalMinutes = config.Opts.SchedulerEntryFrequencyMaxInterval()
		} else {
			intervalMinutes = int(math.Round(float64(7*24*60) / float64(weeklyCount*config.Opts.SchedulerEntryFrequencyFactor())))
			intervalMinutes = int(math.Min(float64(intervalMinutes), float64(config.Opts.SchedulerEntryFrequencyMaxInterval())))
			intervalMinutes = int(math.Max(float64(intervalMinutes), float64(config.Opts.SchedulerEntryFrequencyMinInterval())))
		}
	default:
		intervalMinutes = config.Opts.SchedulerRoundRobinMinInterval()
	}
	// If the feed has a TTL defined, we use it to make sure we don't check it too often.
	if newTTL > intervalMinutes && newTTL > 0 {
		intervalMinutes = newTTL
	}
	f.NextCheckAt = time.Now().Add(time.Minute * time.Duration(intervalMinutes))
}

// FeedCreationRequest represents the request to create a feed.
type FeedCreationRequest struct {
	FeedURL                     string `json:"feed_url"`
	CategoryID                  int64  `json:"category_id"`
	UserAgent                   string `json:"user_agent"`
	Cookie                      string `json:"cookie"`
	Username                    string `json:"username"`
	Password                    string `json:"password"`
	Crawler                     bool   `json:"crawler"`
	Disabled                    bool   `json:"disabled"`
	NoMediaPlayer               bool   `json:"no_media_player"`
	IgnoreHTTPCache             bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates bool   `json:"allow_self_signed_certificates"`
	FetchViaProxy               bool   `json:"fetch_via_proxy"`
	ScraperRules                string `json:"scraper_rules"`
	RewriteRules                string `json:"rewrite_rules"`
	BlocklistRules              string `json:"blocklist_rules"`
	KeeplistRules               string `json:"keeplist_rules"`
	NSFW                        bool   `json:"nsfw"`
	UrlRewriteRules             string `json:"urlrewrite_rules"`
	DisableHTTP2                bool   `json:"disable_http2"`
}

type FeedCreationRequestFromSubscriptionDiscovery struct {
	Content      io.ReadSeeker
	ETag         string
	LastModified string

	FeedCreationRequest
}

// FeedModificationRequest represents the request to update a feed.
type FeedModificationRequest struct {
	FeedURL                     *string `json:"feed_url"`
	SiteURL                     *string `json:"site_url"`
	Title                       *string `json:"title"`
	ScraperRules                *string `json:"scraper_rules"`
	RewriteRules                *string `json:"rewrite_rules"`
	BlocklistRules              *string `json:"blocklist_rules"`
	KeeplistRules               *string `json:"keeplist_rules"`
	UrlRewriteRules             *string `json:"urlrewrite_rules"`
	Crawler                     *bool   `json:"crawler"`
	UserAgent                   *string `json:"user_agent"`
	Cookie                      *string `json:"cookie"`
	Username                    *string `json:"username"`
	Password                    *string `json:"password"`
	CategoryID                  *int64  `json:"category_id"`
	Disabled                    *bool   `json:"disabled"`
	NoMediaPlayer               *bool   `json:"no_media_player"`
	IgnoreHTTPCache             *bool   `json:"ignore_http_cache"`
	AllowSelfSignedCertificates *bool   `json:"allow_self_signed_certificates"`
	FetchViaProxy               *bool   `json:"fetch_via_proxy"`
	NSFW                        *bool   `json:"nsfw"`
	DisableHTTP2                *bool   `json:"disable_http2"`

	View         *string `json:"view"`
	ProxifyMedia *bool   `json:"proxify_media"`
	CacheMedia   *bool   `json:"cache_media"`
}

// Patch updates a feed with modified values.
func (f *FeedModificationRequest) Patch(feed *Feed) {
	if f.FeedURL != nil && *f.FeedURL != "" {
		feed.FeedURL = *f.FeedURL
	}

	if f.SiteURL != nil && *f.SiteURL != "" {
		feed.SiteURL = *f.SiteURL
	}

	if f.Title != nil && *f.Title != "" {
		feed.Title = *f.Title
	}

	if f.ScraperRules != nil {
		feed.ScraperRules = *f.ScraperRules
	}

	if f.RewriteRules != nil {
		feed.RewriteRules = *f.RewriteRules
	}

	if f.KeeplistRules != nil {
		feed.KeeplistRules = *f.KeeplistRules
	}

	if f.UrlRewriteRules != nil {
		feed.UrlRewriteRules = *f.UrlRewriteRules
	}

	if f.BlocklistRules != nil {
		feed.BlocklistRules = *f.BlocklistRules
	}

	if f.Crawler != nil {
		feed.Crawler = *f.Crawler
	}

	if f.UserAgent != nil {
		feed.UserAgent = *f.UserAgent
	}

	if f.Cookie != nil {
		feed.Cookie = *f.Cookie
	}

	if f.Username != nil {
		feed.Username = *f.Username
	}

	if f.Password != nil {
		feed.Password = *f.Password
	}

	if f.CategoryID != nil && *f.CategoryID > 0 {
		feed.Category.ID = *f.CategoryID
	}

	if f.Disabled != nil {
		feed.Disabled = *f.Disabled
	}

	if f.NoMediaPlayer != nil {
		feed.NoMediaPlayer = *f.NoMediaPlayer
	}

	if f.IgnoreHTTPCache != nil {
		feed.IgnoreHTTPCache = *f.IgnoreHTTPCache
	}

	if f.AllowSelfSignedCertificates != nil {
		feed.AllowSelfSignedCertificates = *f.AllowSelfSignedCertificates
	}

	if f.FetchViaProxy != nil {
		feed.FetchViaProxy = *f.FetchViaProxy
	}

	if f.View != nil {
		feed.View = *f.View
	}

	if f.NSFW != nil {
		feed.NSFW = *f.NSFW
	}

	if f.CacheMedia != nil {
		feed.CacheMedia = *f.CacheMedia
	}

	if f.ProxifyMedia != nil {
		feed.ProxifyMedia = *f.ProxifyMedia
	}

	if f.DisableHTTP2 != nil {
		feed.DisableHTTP2 = *f.DisableHTTP2
	}
}

// Feeds is a list of feed
type Feeds []*Feed
