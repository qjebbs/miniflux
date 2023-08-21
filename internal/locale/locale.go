// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package locale // import "miniflux.app/v2/internal/locale"

// AvailableLanguages returns the list of available languages.
func AvailableLanguages() map[string]string {
	return map[string]string{
		"en_US": "English",
		"es_ES": "Español",
		"fr_FR": "Français",
		"de_DE": "Deutsch",
		"pl_PL": "Polski",
		"pt_BR": "Português Brasileiro",
		"zh_CN": "简体中文",
		"nl_NL": "Nederlands",
		"ru_RU": "Русский",
		"it_IT": "Italiano",
		"ja_JP": "日本語",
	}
}
