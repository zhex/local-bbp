package runner

import (
	"context"
	"fmt"
	"github/zhex/bbp/internal/common"
	"github/zhex/bbp/internal/models"
	"time"
)

type Result struct {
	ID          string
	EventName   string
	StepResults map[float32]*StepResult
	Status      string
}

func NewResult(name string) *Result {
	return &Result{
		ID:          common.NewID("r-"),
		EventName:   name,
		StepResults: make(map[float32]*StepResult),
		Status:      "pending",
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
	if (sr.Index - float32(int(sr.Index))) == 0 {
		return fmt.Sprintf("%d", int(sr.Index))
	}
	return fmt.Sprintf("%.1f", sr.Index)
}

func GetResult(ctx context.Context) *Result {
	return ctx.Value("result").(*Result)
}

func WithResult(ctx context.Context, result *Result) context.Context {
	return context.WithValue(ctx, "result", result)
}
