package filesystem // import "miniflux.app/filesystem"

import (
	"path/filepath"

	"miniflux.app/config"
	"miniflux.app/logger"
)

// StorageRoot returns the root directory of file system storage
func StorageRoot() string {
	var err error
	diskRoot := config.Opts.DiskStorageRoot()
	result := diskRoot
	if !filepath.IsAbs(result) {
		result, err = filepath.Abs(result)
		if err != nil {
			logger.Fatal("[Storage:FileSystem] Cannot resolve path %s: %v", diskRoot, err)
		}
	}
	return result
}
