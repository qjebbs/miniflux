// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package ui

import (
	"net/http"

	"github.com/miniflux/miniflux/http/context"
	"github.com/miniflux/miniflux/http/response/html"
	"github.com/miniflux/miniflux/ui/session"
	"github.com/miniflux/miniflux/ui/view"
)

// AddSubscription shows the form to add a new feed.
func (c *Controller) AddSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := context.New(r)
	sess := session.New(c.store, ctx)
	view := view.New(c.tpl, ctx, sess)

	user, err := c.store.UserByID(ctx.UserID())
	if err != nil {
		html.ServerError(w, err)
		return
	}

	categories, err := c.store.Categories(user.ID)
	if err != nil {
		html.ServerError(w, err)
		return
	}

	view.Set("categories", categories)
	view.Set("menu", "feeds")
	view.Set("user", user)
	view.Set("countUnread", c.store.CountUnreadEntries(user.ID))

	html.OK(w, r, view.Render("add_subscription"))
}
