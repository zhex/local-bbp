package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"os"
	"path"
	"time"
)

type Container struct {
	client      *client.Client
	ID          string
	UID         int
	GID         int
	Inputs      *Input
	Network     *Network
	Vol         *volume.Volume
	shareVolume bool
}

func NewContainer(inputs *Input) *Container {
	return &Container{
		client: dockerClient,
		Inputs: inputs,
	}
}

func (c *Container) IsImageExists(ctx context.Context) (bool, error) {
	_, _, err := c.client.ImageInspectWithRaw(ctx, c.Inputs.Image)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Container) Pull(ctx context.Context) error {
	reader, err := c.client.ImagePull(ctx, c.Inputs.Image, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return err
}

func (c *Container) Create(ctx context.Context, net *Network, vol *volume.Volume) error {
	var envs []string
	for k, v := range c.Inputs.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	conf := &container.Config{
		Image: c.Inputs.Image,
		Tty:   true,
		Env:   envs,
	}

	if vol == nil {
		v, err := c.client.VolumeCreate(ctx, volume.CreateOptions{
			Name: fmt.Sprintf("vol_%s", c.Inputs.Name),
		})
		if err != nil {
			return err
		}
		c.Vol = &v
		c.shareVolume = false
	} else {
		c.Vol = vol
		c.shareVolume = true
	}

	hostConf := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Source: c.Vol.Name,
				Target: c.Inputs.WorkDir,
				Type:   mount.TypeVolume,
			},
		},
	}
	plat := &v1.Platform{}
	networkConf := &network.NetworkingConfig{}

	cr, err := c.client.ContainerCreate(ctx, conf, hostConf, networkConf, plat, c.Inputs.Name)
	if err != nil {
		return err
	}
	c.ID = cr.ID

	c.Network = net
	return c.client.NetworkConnect(ctx, net.ID, c.ID, &network.EndpointSettings{})
}

func (c *Container) Start(ctx context.Context) error {
	return c.client.ContainerStart(ctx, c.ID, container.StartOptions{})
}

func (c *Container) Exec(ctx context.Context, workdir string, cmd []string, outputHandler func(reader io.Reader) error) error {
	exec, err := c.client.ContainerExecCreate(ctx, c.ID, container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		WorkingDir:   workdir,
	})
	if err != nil {
		return err
	}
	resp, err := c.client.ContainerExecAttach(ctx, exec.ID, container.ExecAttachOptions{
		Tty: true,
	})
	if err != nil {
		return err
	}

	if outputHandler != nil {
		if err := outputHandler(resp.Reader); err != nil {
			return err
		}
	}

	for {
		inspectResp, err := c.client.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			return err
		}

		if !inspectResp.Running {
			if inspectResp.ExitCode == 0 {
				return nil
			}
			return fmt.Errorf("exitcode '%d': failure", inspectResp.ExitCode)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Container) Destroy(ctx context.Context) error {
	if c.ID == "" {
		return nil
	}
	c.Network = nil

	err := c.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{
		Force: true,
	})
	if err != nil {
		return err
	}

	if !c.shareVolume {
		return c.client.VolumeRemove(ctx, c.Vol.Name, true)
	}
	return nil

}

func (c *Container) GetLogs(ctx context.Context, handler func(reader io.Reader) error) error {
	reader, err := c.client.ContainerLogs(ctx, c.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	if handler != nil {
		return handler(reader)
	}
	return nil
}

func (c *Container) CopyToContainer(ctx context.Context, source, target string, excludePatterns []string) error {
	tarStream, err := archive.TarWithOptions(source, &archive.TarOptions{
		ExcludePatterns: excludePatterns,
	})
	if err != nil {
		return err
	}
	defer tarStream.Close()
	return c.client.CopyToContainer(ctx, c.ID, target, tarStream, container.CopyToContainerOptions{})
}

func (c *Container) CopyToHost(ctx context.Context, source, target string) error {
	reader, _, err := c.client.CopyFromContainer(ctx, c.ID, path.Join(c.Inputs.WorkDir, source))
	if err != nil {
		return err
	}
	defer reader.Close()

	return archive.Untar(reader, target, &archive.TarOptions{
		NoLchown: true,
	})
}

func (c *Container) Wait(ctx context.Context) error {
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
