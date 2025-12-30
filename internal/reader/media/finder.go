package media // import "miniflux.app/v2/internal/reader/media"

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"miniflux.app/v2/internal/config"
	"miniflux.app/v2/internal/urllib"

	"github.com/PuerkitoBio/goquery"
	"miniflux.app/v2/internal/crypto"
	"miniflux.app/v2/internal/model"
)

var queries = []string{
	"img[src]",
}

// URLHash returns the hash of a media url
func URLHash(mediaURL string) string {
	return crypto.SHA256(strings.Trim(mediaURL, " "))
}

// FindMedia try to find the media cache of the URL.
func FindMedia(media *model.Media) error {
	if strings.HasPrefix(media.URL, "data:") {
		return fmt.Errorf("refuse to cache 'data' scheme media")
	}

	slog.Debug("fetching media", slog.String("url", media.URL))
	resp, err := FetchMedia(media, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unable to fetch media: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read downloaded media: %v", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("downloaded media is empty, mediaURL=%s", media.URL)
	}

	media.URLHash = URLHash(media.URL)
	media.MimeType = resp.Header.Get("Content-Type")
	media.Content = body
	media.Size = len(body)
	media.CreatedAt = time.Now()

	return nil
}

// FetchMedia fetches the media from the URL.
func FetchMedia(media *model.Media, r *http.Request) (*http.Response, error) {
	clt := &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: time.Duration(config.Opts.MediaProxyHTTPClientTimeout()) * time.Second,
		},
		Timeout: time.Duration(config.Opts.HTTPClientTimeout()) * time.Second,
	}
	req, err := http.NewRequest("GET", media.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Connection", "close")
	if r != nil {
		forwardedRequestHeader := []string{"Range", "Accept", "Accept-Encoding"}
		for _, requestHeaderName := range forwardedRequestHeader {
			if r.Header.Get(requestHeaderName) != "" {
				req.Header.Add(requestHeaderName, r.Header.Get(requestHeaderName))
			}
		}
	}
	resp, err := clt.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		if media.Referrer != "" {
			req.Header.Add("Referer", media.Referrer)
		}
		resp, err := clt.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
	}

	return resp, nil
}

// ParseDocument parse the entry content and returns media urls of it
func ParseDocument(entry *model.Entry) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(entry.Content))
	if err != nil {
		return nil, fmt.Errorf("unable to read document: %v", err)
	}

	urls := make([]string, 0)
	for _, query := range queries {
		doc.Find(query).Each(func(i int, s *goquery.Selection) {
			href := strings.Trim(s.AttrOr("src", ""), " ")
			if href == "" || strings.HasPrefix(href, "data:") {
				return
			}
			href, err = urllib.AbsoluteURL(entry.URL, href)
			if err != nil {
				return
			}
			urls = append(urls, href)
		})
	}

	return urls, nil
}
