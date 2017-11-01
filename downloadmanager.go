package downloadmanager

import (
	"fmt"
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unknwon/com"
	"github.com/hashicorp/go-getter"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rai-project/utils"
)

func cleanup(s string) string {
	return strings.Replace(
		strings.Replace(s, " ", "_", -1),
		":", "_", -1,
	)
}

type cacheKey struct {
	url            string
	targetFilePath string
}

func (c cacheKey) String() string {
	return fmt.Sprintf("key:%v,path:%v", c.url, c.targetFilePath)
}

func DownloadFile(url, targetFilePath string, opts ...Option) (string, error) {
	options := NewOptions(opts...)

	if url == "" {
		return "", errors.New("invalid empty url")
	}

	cacheKey := cacheKey{url: url, targetFilePath: targetFilePath}.String()

	if options.cache {
		// Get the string associated with the key url from the cache
		if val, found := cache.Get(cacheKey); found {
			s, ok := val.(string)
			if ok && com.IsFile(targetFilePath) {
				return s, nil
			}
		}
	}

	_, err := getter.Detect(url, targetFilePath, getter.Detectors)
	if err != nil {
		return "", err
	}

	targetDir := cleanup(filepath.Dir(targetFilePath))
	if !com.IsDir(targetDir) {
		err := os.MkdirAll(targetDir, 0700)
		if err != nil {
			return "", errors.Wrapf(err, "failed to create %v directory", targetDir)
		}
	}

	// file already exists, but is not in the cache
	if options.cache && com.IsFile(targetFilePath) {
		if options.checkMd5Sum == false {
			cache.Set(cacheKey, targetFilePath, gocache.DefaultExpiration)
			return targetFilePath, nil
		}
		if options.md5Sum != "" {
			if ok, err := utils.MD5Sum.CheckFile(targetFilePath, options.md5Sum); err == nil && ok {
				// Set the value of the key url to targetDir, with the default expiration time
				cache.Set(cacheKey, targetFilePath, gocache.DefaultExpiration)
				return targetFilePath, nil
			}
		}
		os.RemoveAll(targetFilePath)
	}

	log.WithField("url", url).
		WithField("target", targetFilePath).
		Debug("downloading data")

	pwd := targetDir
	if com.IsFile(targetDir) {
		pwd = filepath.Dir(targetDir)
	}

	client := &getter.Client{
		Src:           url,
		Dst:           targetFilePath,
		Pwd:           pwd,
		Mode:          getter.ClientModeFile,
		Decompressors: map[string]getter.Decompressor{}, // do not decompress
	}
	if err := client.Get(); err != nil {
		return "", err
	}

	// validate checksum
	if options.md5Sum != "" && options.checkMd5Sum {
		if ok, err := utils.MD5Sum.CheckFile(targetFilePath, options.md5Sum); !ok {
			os.RemoveAll(targetFilePath)
			return "", err
		}
	}

	if options.cache {
		// Set the value of the key url to targetDir, with the default expiration time
		cache.Set(cacheKey, targetFilePath, gocache.DefaultExpiration)
	}

	return targetFilePath, nil
}

func DownloadInto(url, targetDir string, opts ...Option) (string, error) {
	options := NewOptions(opts...)

	targetDir = cleanup(targetDir)
	if !com.IsDir(targetDir) {
		err := os.MkdirAll(targetDir, 0700)
		if err != nil {
			return "", errors.Wrapf(err, "failed to create %v directory", targetDir)
		}
	}

	urlParsed, err := urlpkg.Parse(url)
	if err != nil {
		return "", errors.Wrapf(err, "unable to parse url %v", url)
	}
	t := filepath.Join(targetDir, filepath.Base(urlParsed.Path))
	filePath, err := DownloadFile(url, t, WithOptions(options))
	if err != nil {
		return "", errors.Wrapf(err, "unable to download url %v into %v", url, t)
	}

	if err := Unarchive(targetDir, filePath); err != nil {
		return "", err
	}

	if options.cache {
		// Set the value of the key url to targetDir, with the default expiration time
		cacheKey := cacheKey{url: url, targetFilePath: filePath}.String()
		cache.Set(cacheKey, filePath, gocache.DefaultExpiration)
	}

	return filePath, nil
}

func Unarchive(targetDir, filePath string) error {
	matchingLen := 0
	unArchiver := ""
	for k := range getter.Decompressors {
		if strings.HasSuffix(filePath, "."+k) && len(k) > matchingLen {
			unArchiver = k
			matchingLen = len(k)
		}
	}
	if decompressor, ok := getter.Decompressors[unArchiver]; ok {
		decompressor.Decompress(targetDir, filePath, true)
	}

	return nil
}
