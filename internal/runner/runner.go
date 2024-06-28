package runner

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/container"
	"github/zhex/bbp/internal/models"
	"os"
	"runtime"
	"strings"
	"time"
)

type Runner struct {
	Plan   *models.Plan
	Config *Config
}

func NewRunner(plan *models.Plan) *Runner {
	return &Runner{Plan: plan, Config: NewConfig()}
}

func (r *Runner) Run(name string) {
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
			WorkDir: r.Config.WorkDir,
			Image:   image,
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
		NewContainerExecTask(c, sr.Index, sr.Step.Script),
	)

	if len(sr.Step.AfterScript) > 0 {
		t = t.Finally(NewContainerExecTask(c, sr.Index, sr.Step.AfterScript))
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
