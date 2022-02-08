// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"encoding/base64"

	"miniflux.app/config"
	"miniflux.app/http/route"
	"miniflux.app/url"

	"github.com/gorilla/mux"
)

// ProxifyURL generates an URL for a proxified resource.
func ProxifyURL(router *mux.Router, link string) string {
	return route.Path(router, "proxy", "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
}

// ShouldProxify tells if a link should prxified.
func ShouldProxify(link string) bool {
	if link == "" {
		return false
	}
	proxyImages := config.Opts.ProxyImages()
	if isDataURL(link) {
		return false
	}
	return proxyImages == "all" || (proxyImages != "none" && !url.IsHTTPS(link))
}
