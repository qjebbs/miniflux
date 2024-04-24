// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/filesystem"
	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/reader/media"
)

func (h *handler) mediaProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the media is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedDigest := request.RouteStringParam(r, "encodedDigest")
	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("no URL provided"))
		return
	}

	decodedDigest, err := base64.URLEncoding.DecodeString(encodedDigest)
	if err != nil {
		html.BadRequest(w, r, errors.New("unable to decode this digest"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("unable to decode this URL"))
		return
	}

	mac := hmac.New(sha256.New, config.Opts.MediaProxyPrivateKey())
	mac.Write(decodedURL)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decodedDigest, expectedMAC) {
		html.Forbidden(w, r)
		return
	}

	u, err := url.Parse(string(decodedURL))
	if err != nil {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if u.Host == "" {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	if !u.IsAbs() {
		html.BadRequest(w, r, errors.New("invalid URL provided"))
		return
	}

	mediaURL := string(decodedURL)
	slog.Debug("MediaProxy: Fetching remote resource",
		slog.String("media_url", mediaURL),
	)
	etag := crypto.HashFromBytes(decodedURL)

	m, err := h.store.MediaByURL(mediaURL)
	if err != nil {
		goto FETCH
	}
	if m.Content != nil {
		slog.Debug(`proxy from database`, slog.String("media_url", mediaURL))
		response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
			b.WithHeader("Content-Type", m.MimeType)
			b.WithBody(m.Content)
			b.WithoutCompression()
			b.Write()
		})
		return
	}

	if m.Cached {
		// cache is located in file system
		var file *os.File
		file, err = filesystem.MediaFileByHash(m.URLHash)
		if err != nil {
			slog.Debug("Unable to fetch media from file system: %s", err)
			goto FETCH
		}
		defer file.Close()
		slog.Debug(`proxy from filesystem`, slog.String("media_url", mediaURL))
		response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
			b.WithHeader("Content-Type", m.MimeType)
			b.WithBody(file)
			b.WithoutCompression()
			b.Write()
		})
		return
	}

FETCH:
	// TODO: apply config
	// clt := &http.Client{
	// 	Transport: &http.Transport{
	// 		IdleConnTimeout: time.Duration(config.Opts.MediaProxyHTTPClientTimeout()) * time.Second,
	// 	},s
	// 	Timeout: time.Duration(config.Opts.MediaProxyHTTPClientTimeout()) * time.Second,
	// }
	slog.Debug(`fetch and proxy`, slog.String("media_url", mediaURL))
	resp, err := media.FetchMedia(m, r)
	if err != nil {
		slog.Error("MediaProxy: Unable to initialize HTTP client",
			slog.String("media_url", mediaURL),
			slog.Any("error", err),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		slog.Warn("MediaProxy: "+http.StatusText(http.StatusRequestedRangeNotSatisfiable),
			slog.String("media_url", mediaURL),
			slog.Int("status_code", resp.StatusCode),
		)
		html.RequestedRangeNotSatisfiable(w, r, resp.Header.Get("Content-Range"))
		return
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		slog.Warn("MediaProxy: Unexpected response status code",
			slog.String("media_url", mediaURL),
			slog.Int("status_code", resp.StatusCode),
		)
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithStatus(resp.StatusCode)
		b.WithHeader("Content-Security-Policy", `default-src 'self'`)
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))
		forwardedResponseHeader := []string{"Content-Encoding", "Content-Type", "Content-Length", "Accept-Ranges", "Content-Range"}
		for _, responseHeaderName := range forwardedResponseHeader {
			if resp.Header.Get(responseHeaderName) != "" {
				b.WithHeader(responseHeaderName, resp.Header.Get(responseHeaderName))
			}
		}
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}
