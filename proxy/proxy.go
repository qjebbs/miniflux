// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"encoding/base64"
	"net/url"
	"path"

	"miniflux.app/config"
	"miniflux.app/http/route"
	murl "miniflux.app/url"

	"github.com/gorilla/mux"
)

// ProxifyURL generates an URL for a proxified resource.
func ProxifyURL(router *mux.Router, link string) string {
	if link != "" {
		proxyImageURL := config.Opts.ProxyImageUrl()

		if proxyImageURL == "" {
			return route.Path(router, "proxy", "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
		}

		proxyURL, err := url.Parse(proxyImageURL)
		if err != nil {
			return ""
		}

		proxyURL.Path = path.Join(proxyURL.Path, base64.URLEncoding.EncodeToString([]byte(link)))
		return proxyURL.String()
	}
	return ""
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
	return proxyImages == "all" || (proxyImages != "none" && !murl.IsHTTPS(link))
}
