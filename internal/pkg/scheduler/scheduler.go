// internal/pkg/scheduler/scheduler.go
package scheduler

import (
	"context"
	"time"
)

type Task interface {
	Name() string
	Execute(ctx context.Context) error
}

type Scheduler interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	// ScheduleTask schedule using cron value
	ScheduleTask(ctx context.Context, task Task, schedule string) error
	// ScheduleInterval schedule using interval
	ScheduleInterval(ctx context.Context, task Task, interval time.Duration) error
}
