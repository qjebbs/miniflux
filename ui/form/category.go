// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/ui/form"

import (
	"net/http"

	"miniflux.app/v2/model"
)

// CategoryForm represents a feed form in the UI
type CategoryForm struct {
	Title string
	View  string
}

// NewCategoryForm returns a new CategoryForm.
func NewCategoryForm(r *http.Request) *CategoryForm {
	view := r.FormValue("view")
	if _, ok := model.Views()[view]; !ok {
		view = model.ViewDefault
	}
	return &CategoryForm{
		Title: r.FormValue("title"),
		View:  view,
	}
}
