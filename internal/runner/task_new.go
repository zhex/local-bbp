package runner

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/container"
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

func NewImagePullTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		exists, err := c.IsImageExists(ctx)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		log.Debug("pulling image")
		return c.Pull(ctx)
	}
}

func NewContainerCreateTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		log.Debug("creating container")
		return c.Create(ctx)
	}
}

func NewContainerStartTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		return c.Start(ctx)
	}
}

func NewContainerCloneTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		log.Debug("prepare workdir ", c.Inputs.WorkDir)
		cmd := []string{"sh", "-ce", fmt.Sprintf("mkdir -p %s", c.Inputs.WorkDir)}
		if err := c.Exec(ctx, "", cmd, nil); err != nil {
			return err
		}
		log.Debug("cloning container")
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

func NewContainerExecTask(c *container.Container, idx float32, cmd []string) Task {
	return func(ctx context.Context) error {
		result := GetResult(ctx)
		stepResult, ok := result.StepResults[idx]

		if len(cmd) == 0 {
			log.Warn("No script to run")
			if ok {
				stepResult.Outputs["script"] = "No script to run"
				stepResult.Status = "success"
			}
			return nil
		}
		log.Debug("executing script")

		cmd = []string{"sh", "-ce", strings.Join(cmd, "\n")}
		err := c.Exec(ctx, c.Inputs.WorkDir, cmd, func(reader io.Reader) error {
			logPath := fmt.Sprintf("%s/logs/%s-%s.log", result.GetOutputPath(), stepResult.GetIdxString(), stepResult.Name)
			file, err := os.Create(logPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, reader); err != nil {
				return err
			}
			return nil
		})

		if ok {
			if err != nil {
				stepResult.Status = "failed"
			} else {
				stepResult.Status = "success"
			}
		}
		return err
	}
}

func NewContainerRemoveTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		log.Debug("removing container")
		return c.Remove(ctx)
	}
}
