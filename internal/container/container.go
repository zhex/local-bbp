package container

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"os"
	"path"
)

type Container struct {
	client *client.Client
	ID     string
	UID    int
	GID    int
	Inputs *Input
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

func (c *Container) Create(ctx context.Context) error {
	conf := &container.Config{
		Image: c.Inputs.Image,
		Tty:   true,
	}

	hostConf := &container.HostConfig{}
	plat := &v1.Platform{}
	networkConf := &network.NetworkingConfig{}

	cr, err := c.client.ContainerCreate(ctx, conf, hostConf, networkConf, plat, c.Inputs.Name)
	if err != nil {
		return err
	}
	c.ID = cr.ID
	return nil
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

	inspectResp, err := c.client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return err
	}

	if inspectResp.ExitCode == 0 {
		return nil
	}
	return fmt.Errorf("exitcode '%d': failure", inspectResp.ExitCode)
}

func (c *Container) Remove(ctx context.Context) error {
	if c.ID == "" {
		return nil
	}
	return c.client.ContainerRemove(ctx, c.ID, container.RemoveOptions{
		Force: true,
	})
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

func (c *Container) SaveArtifacts(ctx context.Context, targetDir string) error {
	if len(c.Inputs.Artifacts) == 0 {
		return nil
	}
	for _, artifact := range c.Inputs.Artifacts {
		reader, _, err := c.client.CopyFromContainer(ctx, c.ID, artifact)
		if err != nil {
			return err
		}

		file := path.Join(targetDir, artifact)
		writer, err := os.Create(file)
		if err != nil {
			_ = reader.Close()
			return err
		}
		_, err = io.Copy(writer, reader)

		_ = reader.Close()
		_ = writer.Close()

		return err
	}

	return nil
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
