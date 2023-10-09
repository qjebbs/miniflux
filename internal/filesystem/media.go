package filesystem // import "miniflux.app/v2/internal/filesystem"

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"miniflux.app/v2/internal/model"
	"miniflux.app/v2/internal/reader/media"
)

// MediaFromCache loads media from disk cache
func MediaFromCache(m *model.Media) error {
	if strings.HasPrefix(m.URL, "data:") {
		return fmt.Errorf("refuse to cache 'data' scheme media")
	}
	if m.URLHash == "" {
		m.URLHash = media.URLHash(m.URL)
	}
	// logger.Debug("[MediaByHash] Fetching media => %s", media.URL)
	if m.URLHash == "" || m.MimeType == "" {
		return fmt.Errorf("unable to load media cache for empty url or mimetype media")
	}

	fi, err := MediaFileByHash(m.URLHash)
	if err != nil {
		return err
	}
	defer fi.Close()
	chunks := make([]byte, 1024, 1024)
	buf := make([]byte, 1024)
	for {
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if 0 == n {
			break
		}
		chunks = append(chunks, buf[:n]...)
	}
	m.Size = len(chunks)
	m.Content = chunks
	return nil
}

// MediaFileByURL returns an *os.File instance from given URL.
func MediaFileByURL(URL string) (fi *os.File, err error) {
	return os.Open(MediaFilePath(media.URLHash(URL)))
}

// MediaFileByHash returns an *os.File instance from given URL hash.
func MediaFileByHash(hash string) (fi *os.File, err error) {
	return os.Open(MediaFilePath(hash))
}

// MediaFilePath returns the media cache file path for given URL hash
func MediaFilePath(hash string) string {
	return filepath.Join(MediaCacheRoot(), hash[0:2], hash[2:4], hash)
}

// MediaCacheRoot returns the root directory of media cache storage
func MediaCacheRoot() string {
	return filepath.Join(StorageRoot(), "media")
}

// SaveMediaFile save given media to file system
func SaveMediaFile(media *model.Media) error {
	if media.URLHash == "" || len(media.URLHash) < 4 {
		return fmt.Errorf("Invalid media url hash: '%s'", media.URLHash)
	}
	fpath := MediaFilePath(media.URLHash)
	exists, err := Exists(fpath)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	fdir := filepath.Dir(fpath)
	exists, err = Exists(fdir)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(fdir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Unable to create media folders: %v", err)
		}
	}
	fi, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer fi.Close()
	if err != nil {
		return fmt.Errorf("Unable to create media file: %v", err)
	}
	_, err = fi.Write(media.Content)
	if err != nil {
		return fmt.Errorf("Unable to write media file: %v", err)
	}
	return nil
}

// RemoveMediaFile removes a media file from file system by given hash
func RemoveMediaFile(hash string) error {
	fpath := MediaFilePath(hash)
	exists, err := Exists(fpath)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.Remove(MediaFilePath(hash))
}

// ExistMediaFile removes a media file from file system by given hash
func ExistMediaFile(hash string) (bool, error) {
	return Exists(MediaFilePath(hash))
}

// Exists tests if a path exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
