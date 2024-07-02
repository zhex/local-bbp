package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/google/uuid"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/docker"
	"github/zhex/bbp/internal/models"
	"io"
	"os"
	"path"
	"strings"
)

func NewTaskChain(tasks ...Task) Task {
	if len(tasks) == 0 {
		return func(ctx context.Context) error {
			return nil
		}
	}
	var t Task
	for _, task := range tasks {
		if t == nil {
			t = task
		} else {
			t = t.Then(task)
		}
	}
	return t
}

func NewParallelTask(size int, tasks ...Task) Task {
	return func(ctx context.Context) error {
		count := len(tasks)
		taskChan := make(chan Task, count)
		errChan := make(chan error, count)

		if size > count {
			size = count
		}

		for i := 0; i < size; i++ {
			go func(work <-chan Task, errs chan<- error, idx int) {
				for task := range work {
					errs <- task(ctx)
				}
			}(taskChan, errChan, i)
		}

		for i := 0; i < count; i++ {
			taskChan <- tasks[i]
		}
		close(taskChan)

		var firstErr error
		for i := 0; i < count; i++ {
			err := <-errChan
			if firstErr == nil {
				firstErr = err
			}
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		return firstErr
	}
}

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
		log.Debugf("pulling image %s", c.Inputs.Image)
		return c.Pull(ctx)
	}
}

func NewContainerCreateTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		log := GetLogger(ctx)
		result := GetResult(ctx)

		netName := fmt.Sprintf("net_%s", c.Inputs.Name)
		log.Debugf("creating network %s", netName)

		net := docker.NewNetwork(netName)
		if err := net.Create(ctx); err != nil {
			return err
		}

		log.Debugf("creating container %s", c.Inputs.Name)
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
					Source: result.Runner.Config.HostDockerCLI,
					Target: "/usr/local/bin/docker",
					Type:   mount.TypeBind,
				},
			)
		}
		return c.Create(ctx, net, true, mounts)
	}
}

func NewContainerStartTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		return c.Start(ctx)
	}
}

func NewCloneTask(c *docker.Container) Task {
	return func(ctx context.Context) error {
		log := GetLogger(ctx)
		log.Debugf("prepare workdir %s", c.Inputs.WorkDir)
		cmd := []string{"sh", "-ce", fmt.Sprintf("mkdir -p %s && sync", c.Inputs.WorkDir)}
		if err := c.Exec(ctx, "", cmd, nil); err != nil {
			return err
		}
		log.Debugf("cloning project code from %s ", c.Inputs.HostDir)
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
		log := GetLogger(ctx)

		if len(scripts) == 0 {
			log.Warn("No script to run")
			sr.Outputs["script"] = "No script to run"
			sr.Status = "success"
			return nil
		}
		log.Debug("executing script")

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
				envs := common.MergeMaps(c.Inputs.Envs, p.Variables)
				var envArgs []string
				for k, _ := range envs {
					envArgs = append(envArgs, fmt.Sprintf("-e %s=\"$%s\"", k, k))
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
			logPath := fmt.Sprintf("%s/logs/%s-%s.log", result.GetOutputPath(), sr.GetIdxString(), sr.Name)
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
			source := path.Join(result.GetOutputPath(), "artifacts", id)
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
		log := GetLogger(ctx)
		result := GetResult(ctx)

		if sr.Step.Artifacts == nil || len(sr.Step.Artifacts.Paths) == 0 {
			return nil
		}

		for _, pattern := range sr.Step.Artifacts.Paths {
			if pattern == "" {
				continue
			}
			id, _ := uuid.NewUUID()
			log.Debugf("saving artifacts: %s (%s)", pattern, id)
			target := path.Join(result.GetOutputPath(), "artifacts", id.String())
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create target directory: %w", err)
			}

			tarName := "artifact.tar"
			err := c.Exec(ctx, c.Inputs.WorkDir, []string{"sh", "-ce", fmt.Sprintf("tar cvf %s %s", tarName, pattern)}, nil)
			if err != nil {
				return fmt.Errorf("failed to create tarball for pattern: %s", pattern)
			}

			err = c.CopyToHost(ctx, tarName, target)
			if err != nil {
				return err
			}

			artifactFile := path.Join(target, tarName)
			err = common.Untar(artifactFile, target)
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
		log := GetLogger(ctx)
		log.Debugf("destroying container %s", c.Inputs.Name)
		net := c.Network
		if err := c.Destroy(ctx); err != nil {
			return nil
		}
		log.Debugf("destroying network %s", net.Name)
		return net.Destroy(ctx)
	}
}
