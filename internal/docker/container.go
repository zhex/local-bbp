package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type Container struct {
	client          *client.Client
	ID              string
	UID             int
	GID             int
	Inputs          *Input
	Network         *Network
	Vol             *volume.Volume
	DockerDaemonVol *volume.Volume
}

func NewContainer(inputs *Input) *Container {
	return &Container{
		client: dockerClient,
		Inputs: inputs,
	}
}

func (c *Container) IsImageExists(ctx context.Context) (bool, error) {
	_, _, err := c.client.ImageInspectWithRaw(ctx, c.Inputs.Image.Name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Container) Pull(ctx context.Context) error {
	reader, err := c.client.ImagePull(ctx, c.Inputs.Image.Name, image.PullOptions{
		RegistryAuth: c.getAuthString(),
	})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return err
}

func (c *Container) Create(ctx context.Context, net *Network, requireVol bool, mounts []mount.Mount) error {
	var envs []string
	for k, v := range c.Inputs.Envs {
		envs = append(envs, fmt.Sprintf("%s=%s", k, v))
	}
	conf := &container.Config{
		Image: c.Inputs.Image.Name,
		Tty:   true,
		Env:   envs,
		User:  fmt.Sprintf("%d", c.Inputs.Image.RunAsUser),
	}

	if c.DockerDaemonVol != nil {

		conf.Healthcheck = &container.HealthConfig{
			Test:        []string{"CMD", "test", "-e", "/var/run/docker.sock"},
			StartPeriod: 1 * time.Second,
			Interval:    1 * time.Second,
		}
	}

	if c.Inputs.NetworkAlias == "build" {
		conf.Entrypoint = c.Inputs.Entrypoint
	}

	if requireVol {
		v, err := c.client.VolumeCreate(ctx, volume.CreateOptions{
			Name: fmt.Sprintf("vol_%s", c.Inputs.Name),
		})
		if err != nil {
			return err
		}
		c.Vol = &v

		mounts = append(mounts, mount.Mount{
			Source: c.Vol.Name,
			Target: c.Inputs.WorkDir,
			Type:   mount.TypeVolume,
		})
	}

	hostConf := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: container.NetworkMode(net.Name),
		Privileged:  true,
	}
	plat := &v1.Platform{}
	networkConf := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			net.Name: {
				Aliases: []string{c.Inputs.NetworkAlias},
			},
		},
	}

	cr, err := c.client.ContainerCreate(ctx, conf, hostConf, networkConf, plat, c.Inputs.Name)
	if err != nil {
		return err
	}
	c.ID = cr.ID

	c.Network = net
	net.AddService(c)
	return c.client.NetworkConnect(ctx, net.ID, c.ID, &network.EndpointSettings{})
}

func (c *Container) Start(ctx context.Context) error {
	err := c.client.ContainerStart(ctx, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	for {
		inspector, err := c.client.ContainerInspect(ctx, c.ID)
		if err != nil {
			return err
		}
		if inspector.State.Running && (inspector.State.Health == nil || inspector.State.Health.Status == "healthy") {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
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
	defer resp.Close()

	done := make(chan int, 1)
	errChan := make(chan error, 1)

	go func() {
		if outputHandler != nil {
			if err := outputHandler(resp.Reader); err != nil {
				errChan <- err
			}
		}
		for {
			inspectResp, err := c.client.ContainerExecInspect(ctx, exec.ID)
			if err != nil {
				errChan <- err
			}
			if !inspectResp.Running {
				if inspectResp.ExitCode == 0 {
					done <- 0
				} else {
					errChan <- fmt.Errorf("exitcode '%d': failure", inspectResp.ExitCode)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-ctx.Done():
		resp.Close()
		return ctx.Err()
	case <-done:
		return nil
	case err := <-errChan:
		return err
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

	if c.Vol != nil {
		return c.client.VolumeRemove(ctx, c.Vol.Name, true)
	}
	if c.DockerDaemonVol != nil {
		return c.client.VolumeRemove(ctx, c.DockerDaemonVol.Name, true)
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
	if err := os.MkdirAll(target, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
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

func (c *Container) getAuthString() string {
	if c.Inputs.Image.AWS != nil {
		auth, err := c.Inputs.Image.AWS.GetAuthData(c.Inputs.Image.Name)
		if err != nil {
			return ""
		}
		decodedToken, err := base64.StdEncoding.DecodeString(*auth.AuthorizationToken)
		if err != nil {
			return ""
		}
		token := strings.TrimPrefix(string(decodedToken), "AWS:")
		authConfig := registry.AuthConfig{
			Username:      "AWS",
			Password:      token,
			ServerAddress: *auth.ProxyEndpoint,
		}
		encodedJSON, _ := json.Marshal(authConfig)
		return base64.URLEncoding.EncodeToString(encodedJSON)

	}

	if c.Inputs.Image.Username == "" || c.Inputs.Image.Password == "" {
		return ""
	}

	authConfig := registry.AuthConfig{
		Username: c.Inputs.Image.Username,
		Password: c.Inputs.Image.Password,
	}
	encodedJSON, _ := json.Marshal(authConfig)
	return base64.URLEncoding.EncodeToString(encodedJSON)
}
