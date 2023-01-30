package filesystem // import "miniflux.app/filesystem"

import (
	"os"
	"path/filepath"

	"miniflux.app/config"
	"miniflux.app/logger"
)

// StorageRoot returns the root directory of file system storage
func StorageRoot() string {
	var err error
	diskRoot := config.Opts.DiskStorageRoot()
	result := diskRoot
	if filepath.HasPrefix(diskRoot, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatal("[Storage:FileSystem] Cannot resolve path %s: %v", diskRoot, err)
		}
		result = filepath.Join(home, diskRoot[2:])
	} else if !filepath.IsAbs(result) {
		result, err = filepath.Abs(result)
		if err != nil {
			logger.Fatal("[Storage:FileSystem] Cannot resolve path %s: %v", diskRoot, err)
		}
	}
	return result
}
