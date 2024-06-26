package runner

import (
	"context"
	"fmt"
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
