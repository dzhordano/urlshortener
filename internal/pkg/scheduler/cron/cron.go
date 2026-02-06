package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/dzhordano/urlshortener/internal/pkg/logger"
	"github.com/dzhordano/urlshortener/internal/pkg/scheduler"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron  *cron.Cron
	log   logger.Logger
	tasks map[string]cron.EntryID
}

func NewCronScheduler(log logger.Logger) scheduler.Scheduler {
	return &Scheduler{
		cron:  cron.New(),
		log:   log,
		tasks: make(map[string]cron.EntryID),
	}
}

func (s *Scheduler) Start(_ context.Context) error {
	s.log.Info("starting cron scheduler")
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop(_ context.Context) error {
	s.log.Info("stopping cron scheduler")
	s.cron.Stop()
	return nil
}

// ScheduleTask enqueues task for execution.
func (s *Scheduler) ScheduleTask(ctx context.Context, task scheduler.Task, schedule string) error {
	entryID, err := s.cron.AddFunc(schedule, func() {
		s.log.Info("executing scheduled task", "task_name", task.Name())

		// timeout for task completion
		//nolint:mnd // TODO Could make timeout for each task configurable.
		taskCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		if err := task.Execute(taskCtx); err != nil {
			s.log.Error("cron task failed", "task_name", task.Name(), "error", err)
		} else {
			s.log.Debug("cron task completed successfully", "task_name", task.Name())
		}
	})
	if err != nil {
		return fmt.Errorf("failed to schedule task %s: %w", task.Name(), err)
	}

	s.tasks[task.Name()] = entryID
	s.log.Info("scheduled cron task", "task_name", task.Name(), "schedule", schedule)
	return nil
}

func (s *Scheduler) ScheduleInterval(ctx context.Context, task scheduler.Task, interval time.Duration) error {
	// convert to cron time
	schedule := fmt.Sprintf("@every %s", interval)
	return s.ScheduleTask(ctx, task, schedule)
}

func (s *Scheduler) RemoveTask(taskName string) {
	if entryID, ok := s.tasks[taskName]; ok {
		s.cron.Remove(entryID)
		delete(s.tasks, taskName)
		s.log.Info("cron task removed", "task_name", taskName)
	}
}
