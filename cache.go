package downloadmanager

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Unknwon/com"

	gocache "github.com/patrickmn/go-cache"
	"github.com/rai-project/config"
	"github.com/rai-project/utils"
)

var (
	cache                       *gocache.Cache
	DefaultCacheExpiration      = 5 * time.Minute
	DefaultCacheCleanupInterval = 10 * time.Minute
	DefaultCacheSaveInterval    = 5 * time.Minute
)

func init() {
	config.AfterInit(func() {
		config.App.Wait()

		cacheDir := filepath.Join(config.App.TempDir, ".cache", "downloadmanager")
		cacheFile := filepath.Join(cacheDir, ".cache")

		// Create a cache with a default expiration time of DefaultCacheExpiration, and which
		// purges expired items every DefaultCacheCleanupInterval
		cache = gocache.New(DefaultCacheExpiration, DefaultCacheCleanupInterval)

		defer utils.Every(DefaultCacheSaveInterval, func() {
			cache.SaveFile(cacheFile)
		})

		if !com.IsDir(cacheDir) {
			if err := os.MkdirAll(cacheDir, 0700); err != nil {
				return
			}
		}
		if com.IsFile(cacheFile) {
			cache.LoadFile(cacheFile)
		}
	})

}
