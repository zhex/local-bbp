package runner

import (
	"context"
	"github/zhex/bbp/internal/models"
	"time"
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

func (r *Result) GetDuration() time.Duration {
	var start, end time.Time
	for _, sr := range r.StepResults {
		if start.IsZero() || sr.StartTime.Before(start) {
			start = sr.StartTime
		}
		if end.IsZero() || sr.EndTime.After(end) {
			end = sr.EndTime
		}
	}
	return end.Sub(start)
}

type StepResult struct {
	Name      string
	Step      *models.Step
	Outputs   map[string]string
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

func GetResult(ctx context.Context) *Result {
	return ctx.Value("result").(*Result)
}

func WithResult(ctx context.Context, result *Result) context.Context {
	return context.WithValue(ctx, "result", result)
}
