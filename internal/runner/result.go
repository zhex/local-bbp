package runner

import (
	"context"
	"fmt"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/models"
	"strings"
	"time"
)

type Result struct {
	ID          string
	EventName   string
	StepResults map[float32]*StepResult
	Status      string
	Runner      *Runner
}

func NewResult(name string, r *Runner) *Result {
	return &Result{
		ID:          common.NewID("r-"),
		EventName:   name,
		StepResults: make(map[float32]*StepResult),
		Status:      "pending",
		Runner:      r,
	}
}

func (r *Result) AddStep(idx float32, name string, step *models.Step) *StepResult {
	sr := &StepResult{
		Index:   idx,
		Name:    name,
		Step:    step,
		Outputs: make(map[string]string),
		Status:  "pending",
	}
	r.StepResults[idx] = sr
	return sr
}

func (r *Result) GetDuration() time.Duration {
	var start, end time.Time
	for _, sr := range r.StepResults {
		if !sr.StartTime.IsZero() && (start.IsZero() || sr.StartTime.Before(start)) {
			start = sr.StartTime
		}
		if !sr.EndTime.IsZero() && (end.IsZero() || sr.EndTime.After(end)) {
			end = sr.EndTime
		}
	}
	return end.Sub(start)
}

func (r *Result) GetOutputPath() string {
	return fmt.Sprintf("%s/%s", r.Runner.Config.OutputDir, r.ID)
}

type StepResult struct {
	Index     float32
	Name      string
	Step      *models.Step
	Outputs   map[string]string
	StartTime time.Time
	EndTime   time.Time
	Status    string
}

func (sr *StepResult) GetIdxString() string {
	rest := sr.Index - float32(int(sr.Index))
	if rest == 0 {
		return fmt.Sprintf("%d", int(sr.Index))
	}
	return strings.Trim(fmt.Sprintf("%f", sr.Index), "0")
}

func GetResult(ctx context.Context) *Result {
	return ctx.Value("result").(*Result)
}

func WithResult(ctx context.Context, result *Result) context.Context {
	return context.WithValue(ctx, "result", result)
}
