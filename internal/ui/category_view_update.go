// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

func (h *handler) updateCategoryView(w http.ResponseWriter, r *http.Request) {
	categoryID := request.RouteInt64Param(r, "categoryID")
	view := r.FormValue("view")
	if err := h.store.UpdateCategoryView(request.UserID(r), categoryID, view); err != nil {
		json.ServerError(w, r, err)
		return
	}

	json.OK(w, r, "OK")
}
