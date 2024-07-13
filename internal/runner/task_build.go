package runner

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/docker"
	"github.com/zhex/local-bbp/internal/models"
	"io"
	"os"
	"path"
	"strings"
	"sync"
)

//go:embed scripts/get-cache-key.sh
var shaCheckScript string

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

		logger.Debugf("creating build container %s", c.Inputs.Name)
		var mounts []mount.Mount
		if sr.Step.Script.HasPipe() || common.Contains(sr.Step.Services, "docker") {
			vol := &volume.Volume{
				Name: fmt.Sprintf("vol_bbp-%s-docker", sr.GetIdxString()),
			}
			c.DockerDaemonVol = vol
			mounts = append(
				mounts,
				mount.Mount{
					Source: vol.Name,
					Target: "/var/run",
					Type:   mount.TypeVolume,
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
		logger := GetLogger(ctx)
		var wg sync.WaitGroup
		for _, svc := range c.Network.Containers {
			if svc == c {
				continue
			}
			wg.Add(1)
			go func(svc *docker.Container) {
				defer wg.Done()
				if err := svc.Start(ctx); err != nil {
					logger.Errorf("failed to start service %s: %s", svc.Inputs.Name, err.Error())
				} else {
					logger.Debugf("service started: %s", svc.Inputs.Name)
				}
			}(svc)
		}
		wg.Wait()
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

func NewContainerDestroyTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		if c.ID == "" {
			return nil
		}
		logger := GetLogger(ctx)
		logger.Debugf("destroying network and containers %s", c.Inputs.Name)
		return c.Network.Destroy(ctx)
	}
}
