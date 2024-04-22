package runner

import (
	"context"
)

type TaskID string

const (
	TaskUpdateStats TaskID = "service.stats.update"
)

type Task interface {
	ID() TaskID
	Run(ctx context.Context) error
}
