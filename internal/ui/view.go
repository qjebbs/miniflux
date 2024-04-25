// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	json_parser "encoding/json"
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

type viewUpdateRequest struct {
	View string `json:"view"`
}

func (h *handler) updateFeedView(w http.ResponseWriter, r *http.Request) {
	feedID := request.RouteInt64Param(r, "feedID")

	var v viewUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&v); err != nil {
		json.BadRequest(w, r, err)
		return
	}
	if err := h.store.UpdateFeedView(request.UserID(r), feedID, v.View); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}

func (h *handler) updateCategoryView(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")

	var v viewUpdateRequest
	if err := json_parser.NewDecoder(r.Body).Decode(&v); err != nil {
		json.BadRequest(w, r, err)
		return
	}
	if err := h.store.UpdateCategoryView(request.UserID(r), categoryID, v.View); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
