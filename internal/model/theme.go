// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model // import "miniflux.app/v2/internal/model"

// Themes returns the list of available themes.
func Themes() map[string]string {
	return map[string]string{
		"light_serif":       "form.prefs.select.theme_light_serif",
		"light_sans_serif":  "form.prefs.select.theme_light_sans_serif",
		"dark_serif":        "form.prefs.select.theme_dark_serif",
		"dark_sans_serif":   "form.prefs.select.theme_dark_sans_serif",
		"system_serif":      "form.prefs.select.theme_system_serif",
		"system_sans_serif": "form.prefs.select.theme_system_sans_serif",
	}
}

// ThemeColor returns the color for the address bar or/and the browser color.
// https://developer.mozilla.org/en-US/docs/Web/Manifest#theme_color
// https://developers.google.com/web/tools/lighthouse/audits/address-bar
// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meta/name/theme-color
func ThemeColor(theme, colorScheme string) string {
	switch theme {
	case "dark_serif", "dark_sans_serif":
		return "#222"
	case "system_serif", "system_sans_serif":
		if colorScheme == "dark" {
			return "#222"
		}

		return "#fff"
	default:
		return "#fff"
	}
}
