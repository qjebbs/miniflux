// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"net/http"

	"miniflux.app/v2/internal/model"
)

// CategoryForm represents a feed form in the UI
type CategoryForm struct {
	Title        string
	HideGlobally bool

	NSFW bool
	View string
}

// NewCategoryForm returns a new CategoryForm.
func NewCategoryForm(r *http.Request) *CategoryForm {
	view := r.FormValue("view")
	if _, ok := model.Views()[view]; !ok {
		view = model.ViewDefault
	}
	return &CategoryForm{
		Title:        r.FormValue("title"),
		HideGlobally: r.FormValue("hide_globally") == "1",

		NSFW: r.FormValue("nsfw") == "1",
		View: view,
	}
}
