// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"time"

	"miniflux.app/config"
	"miniflux.app/crypto"
	"miniflux.app/filesystem"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/logger"
	"miniflux.app/reader/media"
)

func (h *handler) imageProxy(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the image is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedDigest := request.RouteStringParam(r, "encodedDigest")
	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("No URL provided"))
		return
	}

	decodedDigest, err := base64.URLEncoding.DecodeString(encodedDigest)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this Digest"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this URL"))
		return
	}

	mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
	mac.Write(decodedURL)
	expectedMAC := mac.Sum(nil)

	if !hmac.Equal(decodedDigest, expectedMAC) {
		html.Forbidden(w, r)
		return
	}

	imageURL := string(decodedURL)
	etag := crypto.HashFromBytes(decodedURL)

	m, err := h.store.MediaByURL(imageURL)
	if err != nil {
		goto FETCH
	}
	if m.Content != nil {
		logger.Debug(`[Proxy] From database cache, for %q`, imageURL)
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
			logger.Debug("Unable to fetch media from file system: %s", err)
			goto FETCH
		}
		defer file.Close()
		logger.Debug(`[Proxy] From filesystem cache, for %q`, imageURL)
		response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
			b.WithHeader("Content-Type", m.MimeType)
			b.WithBody(file)
			b.WithoutCompression()
			b.Write()
		})
		return
	}

FETCH:
	logger.Debug(`[Proxy] Fetching %q`, imageURL)
	resp, err := media.FetchMedia(m)
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error(`[Proxy] Status Code is %d for URL %q`, resp.StatusCode, imageURL)
		html.NotFound(w, r)
		return
	}

	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Security-Policy", `default-src 'self'`)
		b.WithHeader("Content-Type", resp.Header.Get("Content-Type"))
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}
