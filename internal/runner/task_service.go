package runner

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/mount"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/docker"
)

func NewCreateServicesTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		result := GetResult(ctx)

		if len(sr.Step.Services) == 0 {
			return nil
		}

		fu := NewFieldUpdater(c.Inputs.Envs)
		for _, service := range sr.Step.Services {
			logger.Debugf("creating service: %s", service)
			svc := result.Runner.Plan.Definitions.Services[service]
			if svc == nil {
				return fmt.Errorf("service not found: %s", service)
			}

			inputs := &docker.Input{
				Name:         fmt.Sprintf("bbp-%s-%s", sr.GetIdxString(), service),
				NetworkAlias: service,
				Image:        svc.Image,
				Envs:         common.MergeMaps(fu.UpdateMap(svc.Variables), c.Inputs.Envs),
			}

			var mounts []mount.Mount

			if svc.IsDockerService() {
				inputs.Entrypoint = []string{"dockerd"}

				mounts = append(mounts, mount.Mount{
					Source: c.DockerDaemonVol.Name,
					Target: "/var/run",
					Type:   mount.TypeVolume,
				})
			}

			sc := docker.NewContainer(inputs)
			if err := sc.Create(ctx, c.Network, false, mounts); err != nil {
				return err
			}
		}
		return nil
	}
}
