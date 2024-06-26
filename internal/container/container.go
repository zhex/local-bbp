package container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	image2 "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"math/rand"
	"os"
)

type Container struct {
	client *client.Client
	ID     string
	UID    int
	GID    int
	inputs *Input
}

func NewContainer(inputs *Input) *Container {
	return &Container{
		client: dockerClient,
		inputs: inputs,
	}
}

func (c *Container) IsImageExists(ctx context.Context, image string) (bool, error) {
	_, _, err := c.client.ImageInspectWithRaw(ctx, image)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Container) Pull(ctx context.Context, image string) error {
	reader, err := c.client.ImagePull(ctx, image, image2.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return err
}

func (c *Container) Create(ctx context.Context, image string) error {
	conf := &container.Config{
		Image: image,
		Tty:   true,
	}

	wd, _ := os.Getwd()
	hostConf := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: wd,
				Target: c.inputs.WorkDir,
			},
		},
	}
	plat := &v1.Platform{}
	networkConf := &network.NetworkingConfig{}

	name := fmt.Sprintf("test-%d", rand.Intn(1000))
	cr, err := c.client.ContainerCreate(ctx, conf, hostConf, networkConf, plat, name)
	if err != nil {
		return err
	}
	c.ID = cr.ID
	return nil
}

func (c *Container) Start(ctx context.Context) error {
	return c.client.ContainerStart(ctx, c.ID, container.StartOptions{})
}

func (c *Container) Exec(ctx context.Context, cmd []string) (string, error) {
	exec, err := c.client.ContainerExecCreate(ctx, c.ID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   c.inputs.WorkDir,
	})
	if err != nil {
		return "", err
	}
	resp, err := c.client.ContainerExecAttach(ctx, exec.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(resp.Reader)
	if err != nil {
		return "", err
	}
	resp.Close()

	inspectResp, err := c.client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return "", err
	}

	if inspectResp.ExitCode == 0 {
		return string(data), nil
	}
	return "", fmt.Errorf("exitcode '%d': failure", inspectResp.ExitCode)
}

func (c *Container) Remove(ctx context.Context) error {
	if c.ID == "" {
		return nil
	}
	return c.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{
		Force: true,
	})
}

func (c *Container) wait(ctx context.Context) error {
	statusCh, errCh := c.client.ContainerWait(ctx, c.ID, container.WaitConditionNotRunning)
	var statusCode int64

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("failed to wait for container: %w", err)
		}
	case status := <-statusCh:
		statusCode = status.StatusCode
	}

	if statusCode == 0 {
		return nil
	}
	return fmt.Errorf("exit with `FAILURE`: %v", statusCode)
}
