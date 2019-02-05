// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package scraper // import "miniflux.app/reader/scraper"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"miniflux.app/crypto"
	"miniflux.app/logger"

	"miniflux.app/http/client"
	"miniflux.app/model"
	"miniflux.app/reader/readability"
	"miniflux.app/reader/sanitizer"

	"github.com/PuerkitoBio/goquery"
)

// FetchEntry downloads a web page and returns an Entry.
func FetchEntry(websiteURL, rules, userAgent string) (*model.Entry, error) {
	clt := client.New(websiteURL)
	if userAgent != "" {
		clt.WithUserAgent(userAgent)
	}

	response, err := clt.Get()
	if err != nil {
		return nil, err
	}

	if response.HasServerFailure() {
		return nil, errors.New("scraper: unable to download web page")
	}

	if !isWhitelistedContentType(response.ContentType) {
		return nil, fmt.Errorf("scraper: this resource is not a HTML document (%s)", response.ContentType)
	}

	if err = response.EnsureUnicodeBody(); err != nil {
		return nil, err
	}

	// The entry URL could redirect somewhere else.
	websiteURL = response.EffectiveURL

	if rules == "" {
		rules = getPredefinedScraperRules(websiteURL)
	}

	var title string
	var content string

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	content = buf.String()
	reader := strings.NewReader(content)
	title, err = findTitle(reader)
	if err != nil {
		return nil, err
	}

	reader.Reset(content)
	if rules != "" {
		logger.Debug(`[Scraper] Using rules %q for %q`, rules, websiteURL)
		content, err = scrapContent(reader, rules)
	} else {
		logger.Debug(`[Scraper] Using readability for %q`, websiteURL)
		content, err = readability.ExtractContent(reader)
	}
	content = sanitizer.Sanitize(websiteURL, content)
	entry := &model.Entry{
		URL:     websiteURL,
		Title:   title,
		Content: content,
		Hash:    crypto.Hash(websiteURL),
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
