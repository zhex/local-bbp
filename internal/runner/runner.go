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
}

func New(project string) *Runner {
	c := NewConfig()
	fullPath, _ := filepath.Abs(project)
	c.HostProjectPath = fullPath
	return &Runner{Config: c}
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
	if r.Plan == nil {
		if err := r.LoadPlan(); err != nil {
			log.Fatalf("Error loading plan: %s", err)
		}
	}
	actions := r.getPipeline(name)
	if actions == nil {
		log.Fatalf("No pipeline [%s] found", name)
	}

	ctx := context.Background()

	result := NewResult(name, r)
	ctx = WithResult(ctx, result)

	if err := os.MkdirAll(fmt.Sprintf("%s/logs", result.GetOutputPath()), 0755); err != nil {
		log.Fatalf("Error creating output directory: %s", err)
	}

	if err := os.MkdirAll(fmt.Sprintf("%s/artifacts", result.GetOutputPath()), 0755); err != nil {
		log.Fatalf("Error creating artifacts directory: %s", err)
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
			log.Println("---------------------------------------------")
			log.Println("Pipeline result: ", getColoredStatus(result.Status))
			log.Println("Total Elapsed Time:", result.GetDuration().Round(time.Millisecond).String())
			return nil
		})
		if err := chain(ctx); err != nil {
			log.Fatalf("Error running task: %s", err)
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
			Name:    fmt.Sprintf("bbp_%s_%s", sr.Result.ID, sr.Name),
			Image:   image,
			HostDir: r.Config.HostProjectPath,
			WorkDir: r.Config.WorkDir,
		},
	)

	t := NewTaskChain(
		func(ctx context.Context) error {
			log.Info("Start ", sr.Name)
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

	return t.Finally(NewContainerRemoveTask(c).Then(func(ctx context.Context) error {
		result := GetResult(ctx)
		stepResult, _ := result.StepResults[sr.Index]
		stepResult.EndTime = time.Now()
		d := stepResult.EndTime.Sub(stepResult.StartTime)
		log.Infof("End %s [%s] %s", sr.Name, getColoredStatus(stepResult.Status), common.ColorGrey(d.Round(time.Millisecond).String()))
		return nil
	}))
}

func (r *Runner) getParallelSize() int {
	ncpu := runtime.NumCPU()
	if 1 > ncpu {
		ncpu = 1
	}
	return ncpu
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
