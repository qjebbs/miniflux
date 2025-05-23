package cli

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"miniflux.app/v2/internal/cli/units"
	"miniflux.app/v2/internal/filesystem"
	"miniflux.app/v2/internal/storage"
)

func packMediaCache(s *storage.Storage, args []string) error {
	if len(args) != 1 {
		return errors.New("Usage: miniflux cache-pack /path/to/output.zip")
	}
	f, err := os.Create(args[0])
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()

	cnt, size, maxID, err := s.CachedMediasStat()
	if err != nil {
		return err
	}
	if cnt == 0 {
		return errors.New("No cached media found")
	}
	nDigits := int(math.Log10(float64(maxID)) + 1)

	fmt.Printf("Packaging %d media files with total size of %s.\n", cnt, units.ByteSize(size))
	// fmt.Printf("continue? (y/n): ")
	// input := ""
	// if fmt.Scanf("%s", &input); input != "y" {
	// 	return errors.New("Aborted")
	// }

	const batch = 100
	pos := int64(0)
	for i := int64(0); i < cnt; i += batch {
		medias, err := s.CachedMedias(i, batch)
		if err != nil {
			return err
		}
		for feed, medias := range medias {
			for _, media := range medias {
				pos++
				var br *bufio.Reader
				if len(media.Content) > 0 {
					br = bufio.NewReader(bytes.NewBuffer(media.Content))
				} else {
					file, err := filesystem.MediaFileByHash(media.URLHash)
					if err != nil {
						fmt.Printf("(%d/%d) %s...%s\n", pos, cnt, filename(
							media.ID, nDigits,
							geussExt(media.MimeType, media.URL, nil),
							media.URL, media.CreatedAt,
						), err)
						continue
					}
					br = bufio.NewReader(file)
				}
				filename := filename(
					media.ID, nDigits,
					geussExt(media.MimeType, media.URL, br),
					media.URL, media.CreatedAt,
				)
				f, err := w.Create(filepath.Join(feed, filename))
				if err != nil {
					return err
				}
				_, err = io.Copy(f, br)
				if err != nil {
					return err
				}
				fmt.Printf("(%d/%d) %s...ok\n", pos, cnt, filename)
			}
		}
	}
	return w.Close()
}

func filename(id int64, nDigits int, ext, urlstr string, created time.Time) string {
	simple := fmt.Sprintf("%0*d-%s.%s", nDigits, id, created.Format("2006-01-02"), ext)
	if ext == "" {
		simple = simple[:len(simple)-1]
	}
	if urlstr != "" {
		uri, err := url.Parse(urlstr)
		if err != nil {
			return simple
		}
		base := filepath.Base(uri.Path)
		if strings.Contains(base, ".") {
			base = base[:len(base)-len(filepath.Ext(base))]
		}
		if conv, err := url.PathUnescape(base); err == nil {
			base = conv
		}
		if ext == "" {
			return fmt.Sprintf("%0*d-%s-%s", nDigits, id, created.Format("2006-01-02"), base)
		}
		return fmt.Sprintf("%0*d-%s-%s.%s", nDigits, id, created.Format("2006-01-02"), base, ext)
	}
	return simple
}

func geussExt(contentType string, url string, br *bufio.Reader) string {
	if contentType != "" {
		kind, sub := splitMime(contentType)
		ext := ""
		types.Types.Range(func(ext, value any) bool {
			typ := value.(types.Type)

			if strings.EqualFold(typ.MIME.Type, kind) && strings.EqualFold(typ.MIME.Subtype, sub) {
				ext = typ.Extension
				return false
			}
			return true
		})
		if ext != "" {
			return ext
		}
	}
	if url != "" {
		ext := filepath.Ext(url)
		if ext != "" {
			ext = ext[1:]
		}
		typ := filetype.GetType(ext)
		if typ != types.Unknown {
			return ext
		}
	}
	if br != nil {
		buf, err := br.Peek(256 + 128)
		if err != nil {
			return ""
		}
		typ, _ := filetype.Get(buf)
		if typ != types.Unknown {
			return typ.Extension
		}
	}
	return ""
}

func splitMime(s string) (string, string) {
	x := strings.Split(s, "/")
	if len(x) > 1 {
		return x[0], x[1]
	}
	return x[0], ""
}
