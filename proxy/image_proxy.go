// Copyright 2020 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package proxy // import "miniflux.app/proxy"

import (
	"strings"

	"miniflux.app/config"
	"miniflux.app/model"
	"miniflux.app/reader/sanitizer"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type urlProxyRewriter func(router *mux.Router, url string) string

// ImageProxyRewriter replaces image URLs with internal proxy URLs.
func ImageProxyRewriter(router *mux.Router, feedProxifyImages bool, data string) string {
	return genericImageProxyRewriter(router, ProxifyURL, feedProxifyImages, data)
}

// AbsoluteImageProxyRewriter do the same as ImageProxyRewriter except it uses absolute URLs.
func AbsoluteImageProxyRewriter(router *mux.Router, host string, entry *model.Entry) string {
	proxifyFunction := func(router *mux.Router, url string) string {
		return AbsoluteProxifyURL(router, host, url)
	}
	return genericImageProxyRewriter(router, proxifyFunction, entry.Feed.ProxifyImages || entry.Feed.CacheMedia, entry.Content)
}

func genericImageProxyRewriter(router *mux.Router, proxifyFunction urlProxyRewriter, feedProxifyImage bool, data string) string {
	proxyImages := config.Opts.ProxyImages()
	if proxyImages == "none" && !feedProxifyImage {
		return data
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		if srcAttr, ok := img.Attr("src"); ok {
			if feedProxifyImage || ShouldProxify(srcAttr) {
				img.SetAttr("src", proxifyFunction(router, srcAttr))
			}
		}

		if srcsetAttr, ok := img.Attr("srcset"); ok {
			if feedProxifyImage || ShouldProxify(srcsetAttr) {
				proxifySourceSet(img, router, proxifyFunction, feedProxifyImage, srcsetAttr)
			}
		}
	})

	doc.Find("picture source").Each(func(i int, sourceElement *goquery.Selection) {
		if srcsetAttr, ok := sourceElement.Attr("srcset"); ok {
			if feedProxifyImage || ShouldProxify(srcsetAttr) {
				proxifySourceSet(sourceElement, router, proxifyFunction, feedProxifyImage, srcsetAttr)
			}
		}
	})

	output, err := doc.Find("body").First().Html()
	if err != nil {
		return data
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, proxifyFunction urlProxyRewriter, feedProxifyImage bool, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if feedProxifyImage || ShouldProxify(imageCandidate.ImageURL) {
			imageCandidate.ImageURL = proxifyFunction(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

func isDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}
