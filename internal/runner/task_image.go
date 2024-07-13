package runner

import (
	"context"
	"github.com/zhex/local-bbp/internal/docker"
)

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
