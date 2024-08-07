package runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/zhex/local-bbp/internal/cache"
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/config"
	"github.com/zhex/local-bbp/internal/docker"
	"github.com/zhex/local-bbp/internal/models"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"runtime"
	"time"
)

type Runner struct {
	Plan       *models.Plan
	Config     *config.Config
	Info       *ProjectInfo
	Secrets    map[string]string
	CacheStore *cache.Store
}

func New(projPath string, conf *config.Config, secrets map[string]string) *Runner {
	return &Runner{
		Config:  conf,
		Info:    NewProjInfo(projPath),
		Secrets: secrets,
	}
}

func (r *Runner) LoadPlan() error {
	data, err := os.ReadFile(path.Join(r.Info.Path, "bitbucket-pipelines.yml"))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &r.Plan)
	if err != nil {
		return err
	}
	if r.Plan.Definitions == nil {
		r.Plan.Definitions = &models.Definition{}
	}
	if r.Plan.Definitions.Services == nil {
		r.Plan.Definitions.Services = make(map[string]*models.Service)
	}
	if r.Plan.Definitions.Services["docker"] == nil {
		r.Plan.Definitions.Services["docker"] = &models.Service{
			Image: &models.Image{
				Name: r.Config.DefaultDockerImage,
			},
			Type: "docker",
		}
	}
	r.CacheStore = cache.NewStore(path.Join(r.Config.OutputDir, "cache"), r.Plan.GetCaches())
	return nil
}

func (r *Runner) Run(name string, targetBranch string) {
	ctx := context.Background()

	if r.Plan == nil {
		if err := r.LoadPlan(); err != nil {
			log.Fatalf("Error loading plan: %s", err)
		}
	}

	result := NewResult(name, r)
	ctx = WithResult(ctx, result)

	logger := NewLogger().WithFields(log.Fields{
		"Pipeline": name,
		"ID":       result.ID,
	})
	ctx = WithLogger(ctx, logger)

	actions := r.Plan.GetPipeline(name)
	if actions == nil {
		logger.Fatalf("No pipeline [%s] found", name)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/logs", result.GetResultPath()), 0755); err != nil {
		logger.Fatalf("Error creating output directory: %s", err)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/artifacts", result.GetResultPath()), 0755); err != nil {
		logger.Fatalf("Error creating artifacts directory: %s", err)
	}

	var chain Task

	for i, action := range actions {
		var actionTask Task
		if action.IsParallel() {
			actionTask = r.newParallelTask(action.Parallel, i, result, targetBranch)
		} else if action.IsStage() {
			actionTask = r.newStageTask(action.Stage, i, result, targetBranch)
		} else {
			idx := float32(i + 1)
			sr := result.AddStep(idx, action.Step.GetName(), action.Step)
			actionTask = r.newStepTask(sr, targetBranch)
		}

		if chain == nil {
			chain = actionTask
		} else {
			chain = chain.Then(actionTask)
		}
	}

	if chain != nil {
		chain = WithTimeout(chain, time.Duration(r.Config.MaxPipelineTimeout)*time.Minute)
		chain = chain.Finally(func(ctx context.Context) error {
			for _, sr := range result.StepResults {
				if sr.Status == "failed" {
					result.Status = "failed"
					break
				}
			}
			if result.Status == "pending" {
				result.Status = "success"
			}
			fmt.Print("\n\n")
			logger.Println("Pipeline result: ", getColoredStatus(result.Status))
			logger.Println("Total Elapsed Time:", result.GetDuration().Round(time.Millisecond).String())
			logger.Println("Output Path:", result.GetResultPath())
			return nil
		})
		logger.Infof("Start pipeline: %s", result.EventName)
		if err := chain(ctx); err != nil {
			logger.Fatalf("Error running task: %s", err)
		}
	}
}

func (r *Runner) newParallelTask(parallel *models.Parallel, i int, result *Result, targetBranch string) Task {
	var parallelTasks []Task
	for j, subAction := range parallel.Actions {
		idx := float32(i+1) + float32(j+1)/10
		sr := result.AddStep(idx, subAction.Step.GetName(), subAction.Step)
		parallelTasks = append(parallelTasks, r.newStepTask(sr, targetBranch))
	}
	return ParallelTask(r.getParallelSize(), parallelTasks...)
}

func (r *Runner) newStageTask(stage *models.Stage, i int, result *Result, targetBranch string) Task {
	// no parallel stages
	// no parallel steps in stage
	var stageTasks []Task
	for j, subAction := range stage.Actions {
		idx := float32(i+1) + float32(j+1)/10
		sr := result.AddStep(idx, subAction.Step.GetName(), subAction.Step)
		stageTasks = append(stageTasks, r.newStepTask(sr, targetBranch))
	}

	t := ChainTask(stageTasks...)

	return t.WithCondition(func() bool {
		changedFiles, err := common.GetGitChangedFiles(r.Info.Path, targetBranch)
		if err != nil {
			log.Warnf("Error getting git diff files: %s", err)
			return false
		}
		return stage.MatchCondition(changedFiles)
	})
}

func (r *Runner) newStepTask(sr *StepResult, targetBranch string) Task {
	image := &models.Image{
		Name: r.Config.DefaultImage,
	}
	if r.Plan.HasImage() {
		image = r.Plan.DefaultImage
	}
	if sr.Step.HasImage() {
		image = sr.Step.Image
	}

	envs := common.MergeMaps(r.getEnvs(sr), r.Secrets)
	c := docker.NewContainer(
		&docker.Input{
			Name:         fmt.Sprintf("bbp-%s-%s", sr.Result.ID, sr.GetIdxString()),
			Image:        image,
			NetworkAlias: "build",
			HostDir:      r.Info.Path,
			WorkDir:      r.Config.WorkDir,
			Envs:         envs,
			Entrypoint:   []string{"/bin/sh"},
		},
	)
	image = NewFieldUpdater(envs).UpdateImage(image)

	t := ChainTask(
		NewImagePullTask(c),
		NewContainerCreateTask(c, sr),
		NewCreateServicesTask(c, sr),
		NewContainerStartTask(c),
		NewCloneTask(c),
		NewCachesRestoreTask(c, sr),
		NewDownloadArtifactsTask(c, sr),
		NewScriptTask(c, sr, sr.Step.Script),
		NewSaveArtifactsTask(c, sr),
		NewCachesSaveTask(c, sr),
	)

	if len(sr.Step.AfterScript) > 0 {
		t = t.Finally(NewCmdTask(c, sr, sr.Step.AfterScript))
	}

	t = t.Finally(NewContainerDestroyTask(c))

	timeout := sr.Result.Runner.Config.MaxStepTimeout
	if sr.Step.MaxTime > 0 {
		timeout = sr.Step.MaxTime
	}

	t = WithTimeout(t, time.Duration(timeout)*time.Minute).
		WithCondition(func() bool {
			changedFiles, err := common.GetGitChangedFiles(r.Info.Path, targetBranch)
			if err != nil {
				log.Warnf("Error getting git diff files: %s", err)
				return false
			}
			return sr.Step.MatchCondition(changedFiles)
		})

	return func(ctx context.Context) error {
		ctx = WithLoggerComposeStepResult(ctx, sr)
		logger := GetLogger(ctx)
		logger.Infof("Start step: %s", sr.Name)
		result := GetResult(ctx)
		stepResult, _ := result.StepResults[sr.Index]
		stepResult.StartTime = time.Now()

		err := t(ctx)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			logger.Info("Step timeout")
		}

		d := stepResult.EndTime.Sub(stepResult.StartTime)
		logger.Infof("End step: %s [%s] %s", sr.Name, getColoredStatus(stepResult.Status), common.ColorGrey(d.Round(time.Millisecond).String()))

		return err
	}
}

func (r *Runner) getParallelSize() int {
	ncpu := runtime.NumCPU()
	if 1 > ncpu {
		ncpu = 1
	}
	return ncpu
}

func (r *Runner) getEnvs(sr *StepResult) map[string]string {
	return map[string]string{
		"BITBUCKET_BUILD_NUMBER":        sr.Result.ID,
		"BITBUCKET_BRANCH":              r.Info.BranchName,
		"BITBUCKET_CLONE_DIR":           r.Config.WorkDir,
		"BITBUCKET_COMMIT":              r.Info.CommitID,
		"BITBUCKET_GIT_HTTP_ORIGIN":     "BITBUCKET_GIT_HTTP_ORIGIN",
		"BITBUCKET_GIT_SSH_ORIGIN":      "BITBUCKET_GIT_SSH_ORIGIN",
		"BITBUCKET_PIPELINE_UUID":       uuid.New().String(),
		"BITBUCKET_PROJECT_KEY":         r.Info.Name,
		"BITBUCKET_PROJECT_UUID":        r.Info.ID,
		"BITBUCKET_REPO_FULL_NAME":      path.Base(r.Info.Path),
		"BITBUCKET_REPO_IS_PRIVATE":     "true",
		"BITBUCKET_REPO_OWNER":          r.Info.Owner,
		"BITBUCKET_REPO_OWNER_UUID":     r.Info.OwnerID,
		"BITBUCKET_REPO_SLUG":           r.Info.RepoID,
		"BITBUCKET_REPO_UUID":           r.Info.RepoID,
		"BITBUCKET_SSH_KEY_FILE":        "/opt/atlassian/pipelines/agent/ssh/id_rsa",
		"BITBUCKET_STEP_RUN_NUMBER":     sr.GetIdxString(),
		"BITBUCKET_STEP_TRIGGERER_UUID": r.Info.OwnerID,
		"BITBUCKET_STEP_UUID":           sr.ID.String(),
		"BITBUCKET_WORKSPACE":           r.Info.Name,
		"CI":                            "true",
		"DOCKER_HOST":                   "unix:///var/run/docker.sock",
		"PIPELINES_JWT_TOKEN":           "PIPELINES_JWT_TOKEN",
	}
}

func getColoredStatus(status string) string {
	switch status {
	case "success":
		return common.ColorGreen(status)
	case "failed":
		return common.ColorRed(status)
	default:
		return common.ColorGrey(status)
	}
}
