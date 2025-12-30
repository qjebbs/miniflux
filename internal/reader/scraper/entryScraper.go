// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"bytes"
	"io"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/metric"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/sanitizer"

	"github.com/PuerkitoBio/goquery"
)

// FetchEntry downloads a web page and returns an Entry.
func FetchEntry(user *model.User, websiteURL, rules, userAgent string, cookie string) (*model.Entry, error) {
	requestBuilder := fetcher.NewRequestBuilder()
	if userAgent != "" {
		requestBuilder.WithUserAgent(userAgent, config.Opts.HTTPClientUserAgent())
	}
	if cookie != "" {
		requestBuilder.WithCookie(cookie)
	}

	startTime := time.Now()
	r, err := ScrapeWebsite(
		requestBuilder,
		websiteURL,
		rules,
	)

	if config.Opts.HasMetricsCollector() {
		status := "success"
		if err != nil {
			status = "error"
		}
		metric.ScraperRequestDuration.WithLabelValues(status).Observe(time.Since(startTime).Seconds())
	}

	if err != nil {
		return nil, err
	}

	var title string
	reader := bytes.NewReader(r.Body)
	title, err = findTitle(reader)
	if err != nil {
		return nil, err
	}
	content := sanitizer.SanitizeHTML(websiteURL, r.Content, &sanitizer.SanitizerOptions{OpenLinksInNewTab: user.OpenExternalLinksInNewTab})
	entry := &model.Entry{
		URL:     websiteURL,
		Title:   title,
		Content: content,
		Hash:    crypto.SHA256(websiteURL),
		Date:    time.Now(),
	}

	return entry, nil
}

// findTitle finds title for the html page
func findTitle(page io.Reader) (string, error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", err
	}
	title := ""
	for _, query := range []string{"title", "h1", "h2", "h3", "h4"} {
		document.Find(query).First().Each(func(i int, selection *goquery.Selection) {
			title = selection.Text()
		})
		if title != "" {
			return title, nil
		}
	}
	return "", nil
}
