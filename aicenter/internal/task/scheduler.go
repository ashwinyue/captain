package task

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Task represents a background task
type Task interface {
	Name() string
	Run(ctx context.Context) error
}

// Scheduler manages background tasks
type Scheduler struct {
	mu      sync.Mutex
	tasks   []Task
	logger  *slog.Logger
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewScheduler creates a new task scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:  make([]Task, 0),
		logger: slog.Default().With("component", "task_scheduler"),
	}
}

// RegisterTask adds a task to the scheduler
func (s *Scheduler) RegisterTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, task)
	s.logger.Info("task registered", "task", task.Name())
}

// RunOnce runs all registered tasks once
func (s *Scheduler) RunOnce(ctx context.Context) {
	for _, task := range s.tasks {
		s.logger.Info("running task", "task", task.Name())
		start := time.Now()

		if err := task.Run(ctx); err != nil {
			s.logger.Error("task failed",
				"task", task.Name(),
				"error", err,
				"duration", time.Since(start))
		} else {
			s.logger.Info("task completed",
				"task", task.Name(),
				"duration", time.Since(start))
		}
	}
}

// StartPeriodic starts periodic execution of tasks
func (s *Scheduler) StartPeriodic(interval time.Duration) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Run immediately on start
		s.RunOnce(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.RunOnce(ctx)
			}
		}
	}()

	s.logger.Info("scheduler started", "interval", interval)
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.cancel()
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}
