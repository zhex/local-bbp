package runner

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

func TestTask_Condition(t *testing.T) {
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
	condTask := task1.WithCondition(func() bool {
		return true
	})
	_ = condTask(ctx)
	assert.Equal(t, "1", str)
	condTask = task2.WithCondition(func() bool {
		return false
	})
	_ = condTask(ctx)
	assert.Equal(t, "1", str)
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

func TestNewTaskChain(t *testing.T) {
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
	var task3 Task = func(ctx context.Context) error {
		str += "3"
		return nil
	}
	chain := ChainTask(task1, task2, task3)
	_ = chain(ctx)
	assert.Equal(t, "123", str)
}

func TestNewParallelTask(t *testing.T) {
	ctx := context.Background()
	str := ""
	var task1 Task = func(ctx context.Context) error {
		time.Sleep(50 * time.Millisecond)
		str += "1"
		return nil
	}
	var task2 Task = func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		str += "2"
		return nil
	}
	var task3 Task = func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		str += "3"
		return nil
	}
	parallel := ParallelTask(3, task1, task2, task3)
	_ = parallel(ctx)
	assert.Equal(t, "312", str)
}
