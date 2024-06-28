package runner

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTask_Then(t *testing.T) {
	ctx := context.Background()
	str := ""
	var task1 Task = func(ctx context.Context) error {
		str += "1"
		return nil
	}
	var task2 Task = func(ctx context.Context) error {
		str += "2"
		return nil
	}
	thenTask := task1.Then(task2)
	_ = thenTask(ctx)
	assert.Equal(t, "12", str)
}

func TestTask_Finally(t *testing.T) {
	ctx := context.Background()
	str := ""
	var task1 Task = func(ctx context.Context) error {
		return errors.New("error")
	}
	var task2 Task = func(ctx context.Context) error {
		str += "2"
		return nil
	}
	thenTask := task1.Finally(task2)
	_ = thenTask(ctx)
	assert.Equal(t, "2", str)
}
