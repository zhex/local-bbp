package runner

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
	chain := NewTaskChain(task1, task2, task3)
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
	parallel := NewParallelTask(3, task1, task2, task3)
	_ = parallel(ctx)
	assert.Equal(t, "312", str)
}
