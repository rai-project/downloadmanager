package downloadmanager

import (
	urlpkg "net/url"
	"os"
	"path/filepath"
	"strings"

	context "golang.org/x/net/context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rai-project/tracer"

	"github.com/Unknwon/com"
	"github.com/hashicorp/go-getter"
	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/rai-project/config"
)

func cleanup(s string) string {
	return strings.Replace(s, ":", "_", -1)
}

func DownloadFile(ctx context.Context, url, targetFilePath string) (string, error) {

	if span, newCtx := opentracing.StartSpanFromContext(ctx, "DownloadFile"); span != nil {
		span.SetTag("url", url)
		span.SetTag("traget_file", targetFilePath)
		ctx = newCtx
		defer span.Finish()
	}

	span := tracer.StartSpan("downloading file", opentracing.Tags{
		"url":    url,
		"target": targetFilePath,
	})
	defer span.Finish()

	if url == "" {
		return "", errors.New("invalid empty url")
	}

	// Get the string associated with the key url from the cache
	if val, found := cache.Get(url); found {
		s, ok := val.(string)
		if ok {
			return s, nil
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

	if com.IsFile(targetFilePath) {
		if config.IsDebug {
			log.Debugf("reusing the data in %v", targetFilePath)
			return targetFilePath, nil
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

	// Set the value of the key url to targetDir, with the default expiration time
	cache.Set(url, targetFilePath, gocache.DefaultExpiration)

	return targetFilePath, nil
}

func DownloadInto(ctx context.Context, url, targetDir string) (string, error) {

	if span, newCtx := opentracing.StartSpanFromContext(ctx, "DownloadURL"); span != nil {
		span.SetTag("url", url)
		span.SetTag("traget_dir", targetDir)
		ctx = newCtx
		defer span.Finish()
	}

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
	filePath, err := DownloadFile(ctx, url, t)
	if err != nil {
		return "", errors.Wrapf(err, "unable to download url %v into %v", url, t)
	}

	if err := unarchive(targetDir, filePath); err != nil {
		return "", err
	}

	// Set the value of the key url to targetDir, with the default expiration time
	cache.Set(url, filePath, gocache.DefaultExpiration)

	return filePath, nil
}

func unarchive(targetDir, filePath string) error {
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
