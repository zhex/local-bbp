package runner

import (
	"context"
	"fmt"
	"github.com/zhex/local-bbp/internal/docker"
	"github.com/zhex/local-bbp/internal/models"
	"io"
	"os"
	"strings"
)

func NewCachesRestoreTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		if len(sr.Step.Caches) == 0 {
			return nil
		}
		logger := GetLogger(ctx)
		result := GetResult(ctx)
		cacheStore := result.Runner.CacheStore

		for _, cacheKey := range sr.Step.Caches {
			logger.Debugf("restoring caches: %s", cacheKey)
			cache := cacheStore.Get(cacheKey)
			if cache == nil {
				logger.Debugf("cache not found: %s", cacheKey)
				continue
			}

			hash := getCacheKey(ctx, c, cacheKey, cache)
			if !cacheStore.HasHashPath(cacheKey, hash) {
				logger.Debugf("cache not found: %s: %s", cacheKey, hash)
				continue
			}

			src := cacheStore.GetHashPath(cacheKey, hash)
			if err := c.CopyToContainer(ctx, src, c.Inputs.WorkDir, []string{}); err != nil {
				return fmt.Errorf("failed to restore cache: %s: %s", cacheKey, hash)
			} else {
				logger.Debugf("cache restored: %s: %s", cacheKey, hash)
			}
		}
		return nil
	}
}

func NewCachesSaveTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		if len(sr.Step.Caches) == 0 {
			return nil
		}
		logger := GetLogger(ctx)
		result := GetResult(ctx)
		cacheStore := result.Runner.CacheStore

		for _, cacheKey := range sr.Step.Caches {
			logger.Debugf("saving caches: %s", cacheKey)
			cache := cacheStore.Get(cacheKey)
			if cache == nil {
				logger.Warnf("cache not found: %s", cacheKey)
				continue
			}

			hash := getCacheKey(ctx, c, cacheKey, cache)
			if !cacheStore.HasHashPath(cacheKey, hash) {
				target := cacheStore.GetHashPath(cacheKey, hash)
				if err := c.CopyToHost(ctx, cache.Path, target); err != nil {
					logger.Debugf("failed to save cache: %s: %s", cacheKey, err.Error())
					_ = os.Remove(target)
				} else {
					logger.Debugf("cache saved: %s: %s", cacheKey, hash)
				}
			} else {
				logger.Debugf("skipp cache save, the cache already exists: %s: %s", cacheKey, hash)
			}

		}
		return nil
	}
}

func getCacheKey(ctx context.Context, c *docker.Container, cacheKey string, cache *models.Cache) string {
	logger := GetLogger(ctx)
	var shaKey = ""
	if cache.IsSmartCache() {
		script := strings.Replace(shaCheckScript, "{{patterns}}", strings.Join(cache.Key.Files, " "), 1)
		cmd := []string{"sh", "-ce", script}
		if err := c.Exec(ctx, c.Inputs.WorkDir, cmd, func(reader io.Reader) error {
			data, err := io.ReadAll(reader)
			if err != nil {
				return err
			}
			ret := strings.Trim(string(data), "\r\n")
			if ret == "NONE" {
				return nil
			}
			shaKey = ret
			return nil
		}); err != nil {
			logger.Warnf("failed to check cache: %s", cacheKey)
		}
	} else {
		shaKey = "static"
	}
	return shaKey
}
