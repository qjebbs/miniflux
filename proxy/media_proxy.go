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

// ProxyRewriter replaces media URLs with internal proxy URLs.
func ProxyRewriter(router *mux.Router, data string) string {
	return genericProxyRewriter(router, ProxifyURL, false, data)
}

// ForceProxyRewriter replaces media URLs with internal proxy URLs.
func ForceProxyRewriter(router *mux.Router, data string) string {
	return genericProxyRewriter(router, ProxifyURL, true, data)
}

// AbsoluteProxyRewriter do the same as ProxyRewriter except it uses absolute URLs.
func AbsoluteProxyRewriter(router *mux.Router, host string, entry *model.Entry) string {
	proxifyFunction := func(router *mux.Router, url string) string {
		return AbsoluteProxifyURL(router, host, url)
	}
	return genericProxyRewriter(router, proxifyFunction, entry.Feed.ProxifyMedia || entry.Feed.CacheMedia, entry.Content)
}

func genericProxyRewriter(router *mux.Router, proxifyFunction urlProxyRewriter, foce bool, data string) string {
	proxyOption := config.Opts.ProxyOption()
	if proxyOption == "none" && !foce {
		return data
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return data
	}

	for _, mediaType := range config.Opts.ProxyMediaTypes() {
		switch mediaType {
		case "image":
			doc.Find("img").Each(func(i int, img *goquery.Selection) {
				if srcAttrValue, ok := img.Attr("src"); ok {
					if foce || ShouldProxify(srcAttrValue) {
						img.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}

				if srcsetAttrValue, ok := img.Attr("srcset"); ok {
					proxifySourceSet(img, router, proxifyFunction, foce, srcsetAttrValue)
				}
			})

			doc.Find("picture source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcsetAttrValue, ok := sourceElement.Attr("srcset"); ok {
					proxifySourceSet(sourceElement, router, proxifyFunction, foce, srcsetAttrValue)
				}
			})

		case "audio":
			doc.Find("audio").Each(func(i int, audio *goquery.Selection) {
				if srcAttrValue, ok := audio.Attr("src"); ok {
					if foce || ShouldProxify(srcAttrValue) {
						audio.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

			doc.Find("audio source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcAttrValue, ok := sourceElement.Attr("src"); ok {
					if foce || ShouldProxify(srcAttrValue) {
						sourceElement.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

		case "video":
			doc.Find("video").Each(func(i int, video *goquery.Selection) {
				if srcAttrValue, ok := video.Attr("src"); ok {
					if foce || ShouldProxify(srcAttrValue) {
						video.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})

			doc.Find("video source").Each(func(i int, sourceElement *goquery.Selection) {
				if srcAttrValue, ok := sourceElement.Attr("src"); ok {
					if foce || ShouldProxify(srcAttrValue) {
						sourceElement.SetAttr("src", proxifyFunction(router, srcAttrValue))
					}
				}
			})
		}
	}

	output, err := doc.Find("body").First().Html()
	if err != nil {
		return data
	}

	return output
}

func proxifySourceSet(element *goquery.Selection, router *mux.Router, proxifyFunction urlProxyRewriter, force bool, srcsetAttrValue string) {
	imageCandidates := sanitizer.ParseSrcSetAttribute(srcsetAttrValue)

	for _, imageCandidate := range imageCandidates {
		if force || ShouldProxify(imageCandidate.ImageURL) {
			imageCandidate.ImageURL = proxifyFunction(router, imageCandidate.ImageURL)
		}
	}

	element.SetAttr("srcset", imageCandidates.String())
}

func isDataURL(s string) bool {
	return strings.HasPrefix(s, "data:")
}
