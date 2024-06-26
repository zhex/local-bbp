package runner

import (
	"context"
	"fmt"
	"github/zhex/bbp/internal/models"
)

type Result struct {
	EventName   string
	StepResults map[string]*StepResult
	Status      string
}

func NewResult(name string) *Result {
	return &Result{
		EventName:   name,
		StepResults: make(map[string]*StepResult),
		Status:      "pending",
	}
}

func (r *Result) AddStep(name string, step *models.Step) {
	r.StepResults[name] = &StepResult{
		Name:    name,
		Step:    step,
		Outputs: make(map[string]string),
		Status:  "pending",
	}
}

type StepResult struct {
	Name    string
	Step    *models.Step
	Outputs map[string]string
	Status  string
}

func GetResult(ctx context.Context) *Result {
	return ctx.Value("result").(*Result)
}

func WithResult(ctx context.Context, result *Result) context.Context {
	return context.WithValue(ctx, "result", result)
}

func PrintResult(result *Result) {
	for _, sr := range result.StepResults {
		fmt.Printf("%s [%s]\n", sr.Name, sr.Status)
		for k, v := range sr.Outputs {
			fmt.Printf("  %s\n    %s\n", k, v)
		}
	}
}
