package runner

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/container"
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
	return &Runner{Config: c, Info: newInfo()}
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

	c := container.NewContainer(
		&container.Input{
			Name:    fmt.Sprintf("bbp-%s-%f", sr.Result.ID, sr.Index),
			Image:   image,
			HostDir: r.Config.HostProjectPath,
			WorkDir: r.Config.WorkDir,
			Envs:    r.getEnvs(),
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
		NewContainerCreateTask(c),
		NewContainerStartTask(c),
		NewContainerCloneTask(c),
		NewContainerDownloadArtifactsTask(c, sr),
		NewContainerExecTask(c, sr, sr.Step.Script),
		NewContainerSaveArtifactsTask(c, sr),
	)

	if len(sr.Step.AfterScript) > 0 {
		// fixme - after script log will overwrite the script log
		t = t.Finally(NewContainerExecTask(c, sr, sr.Step.AfterScript))
	}

	t = t.Finally(NewContainerRemoveTask(c).Then(func(ctx context.Context) error {
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

func (r *Runner) getEnvs() map[string]string {
	branch, _ := common.GetGitBranch(r.Config.HostProjectPath)
	commit, _ := common.GetGitCommit(r.Config.HostProjectPath)
	owner, _ := common.GetGitOwner(r.Config.HostProjectPath)
	return map[string]string{
		"BITBUCKET_BUILD_NUMBER":        "",
		"BITBUCKET_BRANCH":              branch,
		"BITBUCKET_CLONE_DIR":           r.Config.WorkDir,
		"BITBUCKET_COMMIT":              commit,
		"BITBUCKET_GIT_HTTP_ORIGIN":     "",
		"BITBUCKET_GIT_SSH_ORIGIN":      "",
		"BITBUCKET_PIPELINE_UUID":       os.Getenv("BITBUCKET_PIPELINE_UUID"),
		"BITBUCKET_PROJECT_KEY":         r.Info.Name,
		"BITBUCKET_PROJECT_UUID":        os.Getenv("BITBUCKET_PROJECT_UUID"),
		"BITBUCKET_REPO_FULL_NAME":      path.Base(r.Config.HostProjectPath),
		"BITBUCKET_REPO_IS_PRIVATE":     "true",
		"BITBUCKET_REPO_OWNER":          owner,
		"BITBUCKET_REPO_OWNER_UUID":     os.Getenv("BITBUCKET_REPO_OWNER_UUID"),
		"BITBUCKET_REPO_SLUG":           os.Getenv("BITBUCKET_REPO_SLUG"),
		"BITBUCKET_REPO_UUID":           os.Getenv("BITBUCKET_REPO_UUID"),
		"BITBUCKET_SSH_KEY_FILE":        "",
		"BITBUCKET_STEP_RUN_NUMBER":     os.Getenv("BITBUCKET_STEP_RUN_NUMBER"),
		"BITBUCKET_STEP_TRIGGERER_UUID": os.Getenv("BITBUCKET_STEP_TRIGGERER_UUID"),
		"BITBUCKET_STEP_UUID":           os.Getenv("BITBUCKET_STEP_UUID"),
		"BITBUCKET_WORKSPACE":           r.Info.Name,
		"CI":                            os.Getenv("CI"),
		"DOCKER_HOST":                   os.Getenv("DOCKER_HOST"),
		"PIPELINES_JWT_TOKEN":           os.Getenv("PIPELINES"),
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
