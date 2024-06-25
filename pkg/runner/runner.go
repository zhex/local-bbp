package runner

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github/zhex/bbp/pkg/container"
	"github/zhex/bbp/pkg/models"
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

	var chain Task

	for _, action := range actions {
		var actionTask Task

		if action.IsParallel() {
			var parallelTasks []Task
			for _, step := range action.Parallel.Actions {
				parallelTasks = append(parallelTasks, r.newStepTask(step.Step))
			}
			actionTask = NewParallelTask(10, parallelTasks...)
		} else {
			actionTask = r.newStepTask(action.Step)
		}

		if chain == nil {
			chain = actionTask
		} else {
			chain = chain.Then(actionTask)
		}
	}

	if chain != nil {
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

func (r *Runner) newStepTask(step *models.Step) Task {
	c := container.NewContainer(
		&container.Input{
			WorkDir: r.config.WorkDir,
		},
	)

	image := r.config.DefaultImage
	if step.Image != "" {
		image = step.Image
	}

	return NewTaskChain(
		NewImagePullTask(c, image),
		NewContainerCreateTask(c, image),
		NewContainerStartTask(c),
		NewContainerExecTask(c, step.Script),
	).Finally(NewContainerRemoveTask(c))
}
