// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"
	"strconv"

	"miniflux.app/errors"
	"miniflux.app/model"
)

// EntryForm represents a feed form in the UI
type EntryForm struct {
	FeedID      int64
	Title       string
	URL         string
	CommentsURL string
	Content     string
	Author      string
}

// ValidateModification validates EntryForm fields
func (e EntryForm) ValidateModification() error {
	if e.Content == "" || e.URL == "" || e.Title == "" || e.FeedID == 0 {
		return errors.NewLocalizedError("error.fields_mandatory")
	}
	return nil
}

// Merge updates the fields of the given feed.
func (e EntryForm) Merge(entry *model.Entry) *model.Entry {
	entry.FeedID = e.FeedID
	entry.Title = e.Title
	entry.URL = e.URL
	entry.CommentsURL = e.CommentsURL
	entry.Content = e.Content
	entry.Author = e.Author

	return entry
}

// NewEntryForm parses the HTTP request and returns a EntryForm
func NewEntryForm(r *http.Request) *EntryForm {
	FeedID, err := strconv.Atoi(r.FormValue("feed_id"))
	if err != nil {
		FeedID = 0
	}

	return &EntryForm{
		FeedID:      int64(FeedID),
		Title:       r.FormValue("title"),
		URL:         r.FormValue("url"),
		CommentsURL: r.FormValue("comments_url"),
		Content:     r.FormValue("content"),
		Author:      r.FormValue("author"),
	}
}
