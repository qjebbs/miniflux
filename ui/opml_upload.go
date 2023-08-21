// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ui // import "miniflux.app/v2/ui"

import (
	"net/http"

	"miniflux.app/v2/config"
	"miniflux.app/v2/http/client"
	"miniflux.app/v2/http/request"
	"miniflux.app/v2/http/response/html"
	"miniflux.app/v2/http/route"
	"miniflux.app/v2/logger"
	"miniflux.app/v2/reader/opml"
	"miniflux.app/v2/ui/session"
	"miniflux.app/v2/ui/view"
)

func (h *handler) uploadOPML(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		logger.Error("[UI:UploadOPML] %v", err)
		html.Redirect(w, r, route.Path(h.router, "import"))
		return
	}
	defer file.Close()

	logger.Debug(
		"[UI:UploadOPML] User #%d uploaded this file: %s (%d bytes)",
		user.ID,
		fileHeader.Filename,
		fileHeader.Size,
	)

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))

	if fileHeader.Size == 0 {
		view.Set("errorMessage", "error.empty_file")
		html.OK(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, file); impErr != nil {
		view.Set("errorMessage", impErr)
		html.OK(w, r, view.Render("import"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}

func (h *handler) fetchOPML(w http.ResponseWriter, r *http.Request) {
	user, err := h.store.UserByID(request.UserID(r))
	if err != nil {
		html.ServerError(w, r, err)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		html.Redirect(w, r, route.Path(h.router, "import"))
		return
	}

	logger.Debug(
		"[UI:FetchOPML] User #%d fetching this URL: %s",
		user.ID,
		url,
	)

	nsfw := request.IsNSFWEnabled(r)
	sess := session.New(h.store, request.SessionID(r))
	view := view.New(h.tpl, r, sess)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", h.store.CountUnreadEntries(user.ID, nsfw))
	view.Set("countErrorFeeds", h.store.CountUserFeedsWithErrors(user.ID, nsfw))

	clt := client.NewClientWithConfig(url, config.Opts)
	resp, err := clt.Get()
	if err != nil {
		view.Set("errorMessage", err)
		html.OK(w, r, view.Render("import"))
		return
	}

	if impErr := opml.NewHandler(h.store).Import(user.ID, resp.Body); impErr != nil {
		view.Set("errorMessage", impErr)
		html.OK(w, r, view.Render("import"))
		return
	}

	html.Redirect(w, r, route.Path(h.router, "feeds"))
}
