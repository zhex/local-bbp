package runner

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/google/uuid"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/docker"
	"github.com/zhex/local-bbp/internal/models"
	"io"
	"os"
	"path"
	"strings"
)

//go:embed scripts/get-cache-key.sh
var shaCheckScript string

func NewImagePullTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		log := GetLogger(ctx)
		exists, err := c.IsImageExists(ctx)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		log.Debugf("pulling image %s", c.Inputs.Image.Name)
		return c.Pull(ctx)
	}
}

func NewContainerCreateTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		result := GetResult(ctx)

		netName := fmt.Sprintf("net_%s", c.Inputs.Name)
		logger.Debugf("creating network %s", netName)

		net := docker.NewNetwork(netName)
		if err := net.Create(ctx); err != nil {
			return err
		}

		logger.Debugf("creating container %s", c.Inputs.Name)
		var mounts []mount.Mount
		if sr.Step.Script.HasPipe() || common.Contains(sr.Step.Services, "docker") {
			mounts = append(
				mounts,
				mount.Mount{
					Source: result.Runner.Config.HostDockerDaemon,
					Target: "/var/run/docker.sock",
					Type:   mount.TypeBind,
				},
				mount.Mount{
					Source: path.Join(result.Runner.Config.ToolDir, common.GetArch(), "docker/docker"),
					Target: "/usr/local/bin/docker",
					Type:   mount.TypeBind,
				},
			)
		}
		return c.Create(ctx, net, false, mounts)
	}
}

func NewContainerStartTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		return c.Start(ctx)
	}
}

func NewCloneTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		logger.Debugf("prepare workdir %s", c.Inputs.WorkDir)
		cmd := []string{"sh", "-ce", fmt.Sprintf("mkdir -p %s && sync", c.Inputs.WorkDir)}
		if err := c.Exec(ctx, "", cmd, nil); err != nil {
			return err
		}
		logger.Debugf("cloning project code from %s ", c.Inputs.HostDir)
		excludePatterns := []string{}
		ignoreFile := path.Join(c.Inputs.HostDir, ".gitignore")
		if common.IsFileExists(ignoreFile) {
			data, err := os.ReadFile(ignoreFile)
			if err != nil {
				return err
			}
			excludePatterns = strings.Split(string(data), "\n")
		}
		return c.CopyToContainer(ctx, c.Inputs.HostDir, c.Inputs.WorkDir, excludePatterns)
	}
}

func NewScriptTask(c *docker.Container, sr *StepResult, scripts models.StepScript) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)

		if len(scripts) == 0 {
			logger.Warn("No script to run")
			sr.Outputs["script"] = "No script to run"
			sr.Status = "success"
			return nil
		}
		logger.Debug("executing script")

		var cmd []string
		for _, script := range scripts {
			if script.Type() == models.ScriptTypeCmd {
				s := script.(*models.CmdScript)
				cmd = append(cmd, s.Cmd)
			} else if script.Type() == models.ScriptTypePipe {
				p := script.(*models.Pipe)

				vols := []string{
					fmt.Sprintf("-v %s:%s", c.Inputs.WorkDir, c.Inputs.WorkDir),
					fmt.Sprintf("-v /usr/local/bin/docker:/usr/local/bin/docker:ro"),
				}
				var envArgs []string
				for k := range c.Inputs.Envs {
					envArgs = append(envArgs, fmt.Sprintf("-e %s=\"$%s\"", k, k))
				}
				for k, v := range p.Variables {
					envArgs = append(envArgs, fmt.Sprintf("-e %s=\"%s\"", k, v))
				}
				dockerCmd := fmt.Sprintf(
					"docker run --rm -w $(pwd) \\\n  %s \\\n  %s \\\n  %s",
					strings.Join(vols, " \\\n  "),
					strings.Join(envArgs, " \\\n  "),
					common.GetPipeImage(p.Pipe),
				)
				cmd = append(cmd, dockerCmd)
			} else {
				return fmt.Errorf("unknown script step type: %v", script)
			}
		}

		err := NewCmdTask(c, sr, cmd)(ctx)

		if err != nil {
			sr.Status = "failed"
		} else {
			sr.Status = "success"
		}
		return err
	}
}

func NewCmdTask(c *docker.Container, sr *StepResult, cmd []string) Task {
	return func(ctx context.Context) error {
		result := GetResult(ctx)

		if len(cmd) == 0 {
			return nil
		}

		var newCmd []string
		for _, c := range cmd {
			newCmd = append(newCmd, fmt.Sprintf("echo '+ %s'", c), c, "echo '\n'")
		}

		cmd = []string{"sh", "-ce", strings.Join(newCmd, "\n")}
		return c.Exec(ctx, c.Inputs.WorkDir, cmd, func(reader io.Reader) error {
			logPath := fmt.Sprintf("%s/logs/%s-%s.log", result.GetResultPath(), sr.GetIdxString(), sr.Name)
			file, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, reader); err != nil {
				return err
			}
			return nil
		})
	}
}

func NewDownloadArtifactsTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		log := GetLogger(ctx)
		result := GetResult(ctx)

		if len(result.Artifacts) == 0 || (sr.Step.Artifacts != nil && !sr.Step.Artifacts.Download) {
			return nil
		}

		for id, pattern := range result.Artifacts {
			log.Debugf("downloading artifacts: %s (%s)", pattern, id)
			source := path.Join(result.GetResultPath(), "artifacts", id)
			err := c.CopyToContainer(ctx, source, c.Inputs.WorkDir, []string{})
			if err != nil {
				return err
			}
		}

		return nil
	}

}

func NewSaveArtifactsTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		result := GetResult(ctx)

		if sr.Step.Artifacts == nil || len(sr.Step.Artifacts.Paths) == 0 {
			return nil
		}

		for _, pattern := range sr.Step.Artifacts.Paths {
			if pattern == "" {
				continue
			}
			id, _ := uuid.NewUUID()
			logger.Debugf("saving artifacts: %s (%s)", pattern, id)

			tarName := "artifact.tar"
			err := c.Exec(ctx, c.Inputs.WorkDir, []string{"sh", "-ce", fmt.Sprintf("tar cvf %s %s", tarName, pattern)}, nil)
			if err != nil {
				return fmt.Errorf("failed to create tarball for pattern: %s", pattern)
			}

			target := path.Join(result.GetResultPath(), "artifacts", id.String())
			err = c.CopyToHost(ctx, tarName, target)
			if err != nil {
				return err
			}

			artifactFile := path.Join(target, tarName)
			err = common.ExtractTarFromFile(artifactFile, target)
			if err != nil {
				return fmt.Errorf("failed to untar artifact: %w", err)
			}

			err = os.Remove(artifactFile)
			if err != nil {
				return err
			}

			err = c.Exec(ctx, c.Inputs.WorkDir, []string{"sh", "-ce", fmt.Sprintf("rm %s", tarName)}, nil)
			if err != nil {
				return err
			}

			result.Artifacts[id.String()] = pattern
		}
		return nil
	}
}

func NewContainerDestroyTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		if c.ID == "" {
			return nil
		}
		logger := GetLogger(ctx)
		logger.Debugf("destroying container %s", c.Inputs.Name)
		net := c.Network
		if err := c.Destroy(ctx); err != nil {
			return nil
		}
		logger.Debugf("destroying network %s", net.Name)
		return net.Destroy(ctx)
	}
}

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
