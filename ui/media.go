// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"miniflux.app/filesystem"
	"miniflux.app/http/client"

	"miniflux.app/crypto"
	"miniflux.app/http/request"
	"miniflux.app/http/response"
	"miniflux.app/http/response/html"
	"miniflux.app/http/response/json"
	"miniflux.app/logger"
)

func (h *handler) getMedia(w http.ResponseWriter, r *http.Request) {
	// If we receive a "If-None-Match" header, we assume the image is already stored in browser cache.
	if r.Header.Get("If-None-Match") != "" {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	encodedURL := request.RouteStringParam(r, "encodedURL")
	if encodedURL == "" {
		html.BadRequest(w, r, errors.New("No URL provided"))
		return
	}

	decodedURL, err := base64.URLEncoding.DecodeString(encodedURL)
	if err != nil {
		html.BadRequest(w, r, errors.New("Unable to decode this URL"))
		return
	}
	imageURL := string(decodedURL)

	userID := request.UserID(r)
	etag := crypto.HashFromBytes(decodedURL)
	media, err := h.store.UserMediaByURL(imageURL, userID)
	if err != nil {
		html.ServerError(w, r, errors.New("Unable to query media cache"))
		return
	}

	if media.Content != nil {
		logger.Debug(`[Proxy] From database cache, for %q`, imageURL)
		response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
			b.WithHeader("Content-Type", media.MimeType)
			b.WithBody(media.Content)
			b.WithoutCompression()
			b.Write()
		})
		return
	}

	if media.Cached {
		// cache is located in file system
		var file *os.File
		file, err = filesystem.MediaFileByHash(media.URLHash)
		if err != nil {
			html.ServerError(w, r, errors.New("Unable to fetch media"))
			return
		}
		defer file.Close()
		logger.Debug(`[Proxy] From filesystem cache, for %q`, imageURL)
		response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
			b.WithHeader("Content-Type", media.MimeType)
			b.WithBody(file)
			b.WithoutCompression()
			b.Write()
		})
		return
	}

	clt := client.New(media.URL)
	clt.WithReferrer(media.Referrer)
	resp, err := clt.Get()
	if err != nil {
		html.ServerError(w, r, err)
		return
	}
	if resp.HasServerFailure() {
		html.ServerError(w, r, fmt.Errorf("unable to download media: status=%d", resp.StatusCode))
		return
	}
	response.New(w, r).WithCaching(etag, 72*time.Hour, func(b *response.Builder) {
		b.WithHeader("Content-Type", media.MimeType)
		b.WithBody(resp.Body)
		b.WithoutCompression()
		b.Write()
	})
}

func (h *handler) toggleEntryMediaCache(w http.ResponseWriter, r *http.Request) {
	entryID := request.RouteInt64Param(r, "entryID")
	if err := h.store.ToggleEntryCache(request.UserID(r), entryID); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}

func (h *handler) removeFeedMediaCache(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")
	if err := h.store.RemoveFeedCaches(request.UserID(r), feedID); err != nil {
		html.ServerError(w, r, err)
		return
	}
}
