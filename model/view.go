// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model // import "miniflux.app/model"

// View types
const (
	ViewDefault = "default"
	ViewList    = "list"
	ViewMasonry = "masonry"
)

// Views returns the list of available views.
func Views() map[string]string {
	return map[string]string{
		"default": "form.prefs.select.view_default",
		"list":    "form.prefs.select.view_list",
		"masonry": "form.prefs.select.view_masonry",
	}
}
