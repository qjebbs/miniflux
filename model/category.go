// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

import "fmt"

// Category represents a feed category.
type Category struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
	FeedCount   int    `json:"nb_feeds,omitempty"`
	View        string `json:"view,omitempty"`
	TotalUnread int    `json:"-"`
}

func (c *Category) String() string {
	return fmt.Sprintf("ID=%d, UserID=%d, Title=%s", c.ID, c.UserID, c.Title)
}

// CategoryRequest represents the request to create or update a category.
type CategoryRequest struct {
	Title string `json:"title"`
	View  string `json:"view"`
}

// Patch updates category fields.
func (cr *CategoryRequest) Patch(category *Category) {
	category.Title = cr.Title
	category.View = cr.View
}

// Categories represents a list of categories.
type Categories []*Category
