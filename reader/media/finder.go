package media // import "miniflux.app/reader/media"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"miniflux.app/config"
	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"miniflux.app/crypto"
	"miniflux.app/logger"
	"miniflux.app/model"
)

var queries = []string{
	"img[src]",
}

// URLHash returns the hash of a media url
func URLHash(mediaURL string) string {
	return crypto.Hash(strings.Trim(mediaURL, " "))
}

// FindMedia try to find the media cache of the URL.
func FindMedia(media *model.Media) error {
	if strings.HasPrefix(media.URL, "data:") {
		return fmt.Errorf("refuse to cache 'data' scheme media")
	}

	logger.Debug("[FindMedia] Fetching media => %s", media.URL)
	resp, err := FetchMedia(media)
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
func FetchMedia(media *model.Media) (*http.Response, error) {
	clt := &http.Client{
		Timeout: time.Duration(config.Opts.HTTPClientTimeout()) * time.Second,
	}
	req, err := http.NewRequest("GET", media.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Connection", "close")
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
			href, err = url.AbsoluteURL(entry.URL, href)
			if err != nil {
				return
			}
			urls = append(urls, href)
		})
	}

	return urls, nil
}
