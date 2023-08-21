// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package proxy // import "miniflux.app/v2/internal/proxy"

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"path"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/http/route"
	"miniflux.app/v2/internal/urllib"

	"github.com/gorilla/mux"
)

// ProxifyURL generates a relative URL for a proxified resource.
func ProxifyURL(router *mux.Router, link string) string {
	if link != "" {
		proxyImageUrl := config.Opts.ProxyUrl()

		if proxyImageUrl == "" {
			mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
			mac.Write([]byte(link))
			digest := mac.Sum(nil)
			return route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
		}

		proxyUrl, err := url.Parse(proxyImageUrl)
		if err != nil {
			return ""
		}

		proxyUrl.Path = path.Join(proxyUrl.Path, base64.URLEncoding.EncodeToString([]byte(link)))
		return proxyUrl.String()
	}
	return ""
}

// AbsoluteProxifyURL generates an absolute URL for a proxified resource.
func AbsoluteProxifyURL(router *mux.Router, host, link string) string {
	if link != "" {
		proxyImageUrl := config.Opts.ProxyUrl()

		if proxyImageUrl == "" {
			mac := hmac.New(sha256.New, config.Opts.ProxyPrivateKey())
			mac.Write([]byte(link))
			digest := mac.Sum(nil)
			path := route.Path(router, "proxy", "encodedDigest", base64.URLEncoding.EncodeToString(digest), "encodedURL", base64.URLEncoding.EncodeToString([]byte(link)))
			if config.Opts.HTTPS {
				return "https://" + host + path
			} else {
				return "http://" + host + path
			}
		}

		proxyURL, err := url.Parse(proxyImageUrl)
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
	proxyOption := config.Opts.ProxyOption()
	if isDataURL(link) {
		return false
	}
	return proxyOption == "all" || (proxyOption != "none" && !urllib.IsHTTPS(link))
}
