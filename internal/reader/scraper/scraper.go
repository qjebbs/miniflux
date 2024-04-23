// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scraper // import "miniflux.app/v2/internal/reader/scraper"

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/reader/encoding"
	"miniflux.app/v2/internal/reader/fetcher"
	"miniflux.app/v2/internal/reader/readability"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
)

// ScrapeWebsiteResult represents the result of a website scraping.
type ScrapeWebsiteResult struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Content    string
}

func ScrapeWebsite(requestBuilder *fetcher.RequestBuilder, websiteURL, rules string) (*ScrapeWebsiteResult, error) {
	resp, reqErr := requestBuilder.ExecuteRequest(websiteURL)
	responseHandler := fetcher.NewResponseHandler(resp, reqErr)
	defer responseHandler.Close()

	if localizedError := responseHandler.LocalizedError(); localizedError != nil {
		slog.Warn("Unable to scrape website", slog.String("website_url", websiteURL), slog.Any("error", localizedError.Error()))
		return nil, localizedError.Error()
	}

	if !isAllowedContentType(responseHandler.ContentType()) {
		return nil, fmt.Errorf("scraper: this resource is not a HTML document (%s)", responseHandler.ContentType())
	}

	// The entry URL could redirect somewhere else.
	sameSite := urllib.Domain(websiteURL) == urllib.Domain(responseHandler.EffectiveURL())
	websiteURL = responseHandler.EffectiveURL()

	if rules == "" {
		rules = getPredefinedScraperRules(websiteURL)
	}

	var content string
	var err error

	htmlDocumentReader, err := encoding.CharsetReaderFromContentType(
		responseHandler.ContentType(),
		responseHandler.Body(config.Opts.HTTPClientMaxBodySize()),
	)
	if err != nil {
		return nil, fmt.Errorf("scraper: unable to read HTML document: %v", err)
	}
	body, err := io.ReadAll(htmlDocumentReader)
	if err != nil {
		return nil, fmt.Errorf("scraper: unable to read HTML document: %v", err)
	}
	htmlDocumentReader = strings.NewReader(string(body))

	if sameSite && rules != "" {
		slog.Debug("Extracting content with custom rules",
			"url", websiteURL,
			"rules", rules,
		)
		content, err = findContentUsingCustomRules(htmlDocumentReader, rules)
	} else {
		slog.Debug("Extracting content with readability",
			"url", websiteURL,
		)
		content, err = readability.ExtractContent(htmlDocumentReader)
	}

	if err != nil {
		return nil, err
	}

	return &ScrapeWebsiteResult{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
		Content:    content,
	}, nil
}

func findContentUsingCustomRules(page io.Reader, rules string) (string, error) {
	document, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return "", err
	}

	contents := ""
	document.Find(rules).Each(func(i int, s *goquery.Selection) {
		var content string

		content, _ = goquery.OuterHtml(s)
		contents += content
	})

	return contents, nil
}

func getPredefinedScraperRules(websiteURL string) string {
	urlDomain := urllib.Domain(websiteURL)

	for domain, rules := range predefinedRules {
		if strings.Contains(urlDomain, domain) {
			return rules
		}
	}

	return ""
}

func isAllowedContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/html") ||
		strings.HasPrefix(contentType, "application/xhtml+xml")
}
