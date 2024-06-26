package runner

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/internal/container"
	"github/zhex/bbp/internal/models"
	"runtime"
	"strings"
)

type Runner struct {
	plan   *models.Plan
	config *Config
}

func NewRunner(plan *models.Plan) *Runner {
	return &Runner{plan: plan, config: NewConfig()}
}

func (r *Runner) Run(name string) {
	actions := r.getPipeline(name)
	if actions == nil {
		log.Fatalf("No pipeline [%s] found", name)
	}

	ctx := context.Background()

	result := NewResult(name)
	ctx = WithResult(ctx, result)

	var chain Task

	for i, action := range actions {
		var actionTask Task

		if action.IsParallel() {
			var parallelTasks []Task
			for j, subAction := range action.Parallel.Actions {
				name := fmt.Sprintf("%d.%d - %s", i+1, j+1, subAction.Step.GetName())
				result.AddStep(name, subAction.Step)
				parallelTasks = append(parallelTasks, r.newStepTask(name, subAction.Step))
			}
			actionTask = NewParallelTask(r.getParallelSize(), parallelTasks...)
		} else {
			name := fmt.Sprintf("%d - %s", i+1, action.Step.GetName())
			result.AddStep(name, action.Step)
			actionTask = r.newStepTask(name, action.Step)
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
			PrintResult(result)
			return nil
		})
		if err := chain(ctx); err != nil {
			log.Fatalf("Error running task: %s", err)
		}
	}
}

func (r *Runner) getPipeline(name string) []*models.Action {
	if strings.ToLower(name) == "default" || len(name) == 0 {
		return r.plan.Pipelines.Default
	}
	if _, ok := r.plan.Pipelines.Custom[name]; ok {
		return r.plan.Pipelines.Custom[name]
	}
	return nil
}

func (r *Runner) newStepTask(name string, step *models.Step) Task {
	c := container.NewContainer(
		&container.Input{
			WorkDir: r.config.WorkDir,
		},
	)

	image := r.config.DefaultImage
	if step.Image != "" {
		image = step.Image
	}

	t := NewTaskChain(
		NewImagePullTask(c, image),
		NewContainerCreateTask(c, image),
		NewContainerStartTask(c),
		NewContainerExecTask(c, name, step.Script),
	)

	if len(step.AfterScript) > 0 {
		t = t.Finally(NewContainerExecTask(c, name, step.AfterScript))
	}

	return t.Finally(NewContainerRemoveTask(c))
}

func (r *Runner) getParallelSize() int {
	ncpu := runtime.NumCPU()
	if 1 > ncpu {
		ncpu = 1
	}
	return ncpu
}
