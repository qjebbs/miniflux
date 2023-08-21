// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/html"
	"miniflux.app/v2/internal/http/response/json"
)

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
