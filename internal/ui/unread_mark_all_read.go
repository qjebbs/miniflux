// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/internal/ui"

import (
	"net/http"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response/json"
)

func (h *handler) markAllAsRead(w http.ResponseWriter, r *http.Request) {
	var err error
	if request.IsNSFWEnabled(r) {
		err = h.store.MarkAllAsReadExceptNSFW(request.UserID(r))
	} else {
		err = h.store.MarkAllAsRead(request.UserID(r))
	}

	if err != nil {
		json.ServerError(w, r, err)
		return
	}
	json.OK(w, r, "OK")
}
