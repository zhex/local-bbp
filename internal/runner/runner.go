package runner

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/docker"
	"github/zhex/bbp/internal/models"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Runner struct {
	Plan   *models.Plan
	Config *Config
	Info   *info
}

func New(project string) *Runner {
	c := NewConfig()
	fullPath, _ := filepath.Abs(project)
	c.HostProjectPath = fullPath
	return &Runner{Config: c, Info: newInfo(fullPath)}
}

func (r *Runner) LoadPlan() error {
	data, err := os.ReadFile(path.Join(r.Config.HostProjectPath, "bitbucket-pipelines.yml"))
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &r.Plan)
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) Run(name string) {
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

	actions := r.getPipeline(name)
	if actions == nil {
		logger.Fatalf("No pipeline [%s] found", name)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/logs", result.GetOutputPath()), 0755); err != nil {
		logger.Fatalf("Error creating output directory: %s", err)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/artifacts", result.GetOutputPath()), 0755); err != nil {
		logger.Fatalf("Error creating artifacts directory: %s", err)
	}

	var chain Task

	for i, action := range actions {
		var actionTask Task

		if action.IsParallel() {
			var parallelTasks []Task
			for j, subAction := range action.Parallel.Actions {
				idx := float32(i+1) + float32(j+1)/10
				sr := result.AddStep(idx, subAction.Step.GetName(), subAction.Step)
				parallelTasks = append(parallelTasks, r.newStepTask(sr))
			}
			actionTask = NewParallelTask(r.getParallelSize(), parallelTasks...)
		} else {
			idx := float32(i + 1)
			sr := result.AddStep(idx, action.Step.GetName(), action.Step)
			actionTask = r.newStepTask(sr)
		}

		if chain == nil {
			chain = actionTask
		} else {
			chain = chain.Then(actionTask)
		}
	}

	if chain != nil {
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
			logger.Println("Output Path:", result.GetOutputPath())
			return nil
		})
		logger.Infof("Start pipeline: %s", result.EventName)
		if err := chain(ctx); err != nil {
			logger.Fatalf("Error running task: %s", err)
		}
	}
}

func (r *Runner) getPipeline(name string) []*models.Action {
	if strings.ToLower(name) == "default" || len(name) == 0 {
		return r.Plan.Pipelines.Default
	}
	if _, ok := r.Plan.Pipelines.Custom[name]; ok {
		return r.Plan.Pipelines.Custom[name]
	}
	return nil
}

func (r *Runner) newStepTask(sr *StepResult) Task {
	image := r.Config.DefaultImage
	if sr.Step.Image != "" {
		image = sr.Step.Image
	}

	c := docker.NewContainer(
		&docker.Input{
			Name:    fmt.Sprintf("bbp-%s-%s", sr.Result.ID, sr.GetIdxString()),
			Image:   image,
			HostDir: r.Config.HostProjectPath,
			WorkDir: r.Config.WorkDir,
			Envs:    r.getEnvs(sr),
		},
	)

	t := NewTaskChain(
		func(ctx context.Context) error {
			ctx = WithLoggerComposeStepResult(ctx, sr)
			logger := GetLogger(ctx)
			logger.Infof("Start step: %s", sr.Name)
			result := GetResult(ctx)
			stepResult, _ := result.StepResults[sr.Index]
			stepResult.StartTime = time.Now()
			return nil
		},
		NewImagePullTask(c),
		NewContainerCreateTask(c, sr),
		NewContainerStartTask(c),
		NewCloneTask(c),
		NewDownloadArtifactsTask(c, sr),
		NewScriptTask(c, sr, sr.Step.Script),
		NewSaveArtifactsTask(c, sr),
	)

	if len(sr.Step.AfterScript) > 0 {
		t = t.Finally(NewCmdTask(c, sr, sr.Step.AfterScript))
	}

	t = t.Finally(NewContainerDestroyTask(c).Then(func(ctx context.Context) error {
		logger := GetLogger(ctx)
		result := GetResult(ctx)
		stepResult, _ := result.StepResults[sr.Index]
		stepResult.EndTime = time.Now()
		d := stepResult.EndTime.Sub(stepResult.StartTime)
		logger.Infof("End step: %s [%s] %s", sr.Name, getColoredStatus(stepResult.Status), common.ColorGrey(d.Round(time.Millisecond).String()))
		return nil
	}))

	return func(ctx context.Context) error {
		return t(WithLoggerComposeStepResult(ctx, sr))
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
		"BITBUCKET_PROJECT_KEY":         r.Info.ProjectName,
		"BITBUCKET_PROJECT_UUID":        r.Info.ProjectID,
		"BITBUCKET_REPO_FULL_NAME":      path.Base(r.Config.HostProjectPath),
		"BITBUCKET_REPO_IS_PRIVATE":     "true",
		"BITBUCKET_REPO_OWNER":          r.Info.Owner,
		"BITBUCKET_REPO_OWNER_UUID":     r.Info.OwnerID,
		"BITBUCKET_REPO_SLUG":           r.Info.RepoID,
		"BITBUCKET_REPO_UUID":           r.Info.RepoID,
		"BITBUCKET_SSH_KEY_FILE":        "/opt/atlassian/pipelines/agent/ssh/id_rsa",
		"BITBUCKET_STEP_RUN_NUMBER":     sr.GetIdxString(),
		"BITBUCKET_STEP_TRIGGERER_UUID": r.Info.OwnerID,
		"BITBUCKET_STEP_UUID":           sr.ID.String(),
		"BITBUCKET_WORKSPACE":           r.Info.ProjectName,
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
