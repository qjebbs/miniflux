package media // import "miniflux.app/reader/media"

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"miniflux.app/url"

	"github.com/PuerkitoBio/goquery"
	"miniflux.app/crypto"
	"miniflux.app/http/client"
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
	return downloadMedia(media)
}

func downloadMedia(media *model.Media) error {
	clt := client.New(media.URL)
	response, err := clt.Get()
	if err != nil {
		return fmt.Errorf("unable to download mediaURL: %v", err)
	}

	if response.HasServerFailure() && media.Referrer != "" {
		clt.WithReferrer(media.Referrer)
		response, err = clt.Get()
		if err != nil {
			return fmt.Errorf("unable to download mediaURL: %v", err)
		}
	}

	if response.HasServerFailure() {
		return fmt.Errorf("unable to download media: status=%d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read downloaded media: %v", err)
	}

	if len(body) == 0 {
		return fmt.Errorf("downloaded media is empty, mediaURL=%s", media.URL)
	}

	media.URLHash = URLHash(media.URL)
	media.MimeType = response.ContentType
	media.Content = body
	media.Size = len(body)
	media.CreatedAt = time.Now()

	return nil
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
