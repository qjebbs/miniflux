// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package form // import "miniflux.app/ui/form"

import (
	"net/http"

	"miniflux.app/model"
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
