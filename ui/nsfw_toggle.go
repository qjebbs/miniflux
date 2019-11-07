// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui // import "miniflux.app/ui"

import (
	"net/http"

	"miniflux.app/http/request"
	"miniflux.app/http/response/json"
	"miniflux.app/ui/session"
)

func (h *handler) toggleNSFW(w http.ResponseWriter, r *http.Request) {
	sess := session.New(h.store, request.SessionID(r))
	if request.IsNSFWEnabled(r) {
		sess.SetNSFW("show")
	} else {
		sess.SetNSFW("hide")
	}
	json.OK(w, r, "OK")
}
