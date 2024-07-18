package runner

import (
	"context"
	"fmt"
	"time"
)

type Task func(ctx context.Context) error

func (t Task) Then(next Task) Task {
	return func(ctx context.Context) error {
		err := t(ctx)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return next(ctx)
	}
}

func (t Task) WithCondition(cond func() bool) Task {
	return func(ctx context.Context) error {
		logger := GetLogger(ctx)
		if !cond() {
			logger.Infof("Skip step")
			return nil
		}
		return t(ctx)
	}
}

func (t Task) Finally(ft Task) Task {
	return func(ctx context.Context) error {
		err := t(ctx)
		errf := ft(ctx)
		if errf != nil {
			return fmt.Errorf("error in task finally: %w", errf)
		}
		return err
	}
}

func ChainTask(tasks ...Task) Task {
	if len(tasks) == 0 {
		return func(ctx context.Context) error {
			return nil
		}
	}
	var t Task
	for _, task := range tasks {
		if t == nil {
			t = task
		} else {
			t = t.Then(task)
		}
	}
	return t
}

func ParallelTask(size int, tasks ...Task) Task {
	return func(ctx context.Context) error {
		count := len(tasks)
		taskChan := make(chan Task, count)
		errChan := make(chan error, count)

		if size > count {
			size = count
		}

		for i := 0; i < size; i++ {
			go func(work <-chan Task, errs chan<- error, idx int) {
				for task := range work {
					errs <- task(ctx)
				}
			}(taskChan, errChan, i)
		}

		for i := 0; i < count; i++ {
			taskChan <- tasks[i]
		}
		close(taskChan)

		var firstErr error
		for i := 0; i < count; i++ {
			err := <-errChan
			if firstErr == nil {
				firstErr = err
			}
		}

		if err := ctx.Err(); err != nil {
			return err
		}
		return firstErr
	}
}

func WithTimeout(task Task, timeout time.Duration) Task {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return task(ctx)
	}
}
