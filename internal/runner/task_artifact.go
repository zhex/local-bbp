package runner

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/docker"
	"os"
	"path"
)

func NewSaveArtifactsTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		result := GetResult(ctx)

		if sr.Step.Artifacts == nil || len(sr.Step.Artifacts.Paths) == 0 {
			return nil
		}

		for _, pattern := range sr.Step.Artifacts.Paths {
			if pattern == "" {
				continue
			}
			id, _ := uuid.NewUUID()
			logger.Debugf("saving artifacts: %s (%s)", pattern, id)

			tarName := "artifact.tar"
			err := c.Exec(ctx, c.Inputs.WorkDir, []string{"sh", "-ce", fmt.Sprintf("tar cvf %s %s", tarName, pattern)}, nil)
			if err != nil {
				return fmt.Errorf("failed to create tarball for pattern: %s", pattern)
			}

			target := path.Join(result.GetResultPath(), "artifacts", id.String())
			err = c.CopyToHost(ctx, tarName, target)
			if err != nil {
				return err
			}

			artifactFile := path.Join(target, tarName)
			err = common.ExtractTarFromFile(artifactFile, target)
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

func NewDownloadArtifactsTask(c *docker.Container, sr *StepResult) Task {
	return func(ctx context.Context) error {
		log := GetLogger(ctx)
		result := GetResult(ctx)

		if len(result.Artifacts) == 0 || (sr.Step.Artifacts != nil && !sr.Step.Artifacts.Download) {
			return nil
		}

		for id, pattern := range result.Artifacts {
			log.Debugf("downloading artifacts: %s (%s)", pattern, id)
			source := path.Join(result.GetResultPath(), "artifacts", id)
			err := c.CopyToContainer(ctx, source, c.Inputs.WorkDir, []string{})
			if err != nil {
				return err
			}
		}

		return nil
	}

}
