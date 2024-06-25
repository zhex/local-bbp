package runner

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/pkg/container"
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

func NewImagePullTask(c *container.Container, image string) Task {
	return func(ctx context.Context) error {
		exists, err := c.IsImageExists(ctx, image)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
		log.Debug("pulling image")
		return c.Pull(ctx, image)
	}
}

func NewContainerCreateTask(c *container.Container, image string) Task {
	return func(ctx context.Context) error {
		log.Debug("creating container")
		return c.Create(ctx, image)
	}
}

func NewContainerStartTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		return c.Start(ctx)
	}
}

func NewContainerExecTask(c *container.Container, cmd []string) Task {
	return func(ctx context.Context) error {
		if len(cmd) == 0 {
			log.Warn("No script to run")
			return nil
		}
		log.Debug("executing script")
		return c.Exec(ctx, []string{"sh", "-ce", strings.Join(cmd, "\n")})
	}
}

func NewContainerRemoveTask(c *container.Container) Task {
	return func(ctx context.Context) error {
		log.Debug("removing container")
		return c.Remove(ctx)
	}
}
