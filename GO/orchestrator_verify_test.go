package main

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusRunning   TaskStatus = "RUNNING"
	StatusCompleted TaskStatus = "COMPLETED"
	StatusFailed    TaskStatus = "FAILED"
)

type TaskFn func(ctx context.Context) error

type Task struct {
	ID           string
	Execute      TaskFn
	Dependencies []string
	Status       TaskStatus
	Err          error
}

type DAGOrchestrator struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewDAGOrchestrator() *DAGOrchestrator {
	return &DAGOrchestrator{tasks: make(map[string]*Task)}
}

func (o *DAGOrchestrator) AddTask(id string, deps []string, fn TaskFn) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tasks[id] = &Task{
		ID:           id,
		Execute:      fn,
		Dependencies: deps,
		Status:       StatusPending,
	}
}

func (o *DAGOrchestrator) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	completed := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(o.tasks))

	for {
		o.mu.Lock()
		allDone := true
		readyTasks := make([]*Task, 0)

		for _, task := range o.tasks {
			if task.Status != StatusCompleted && task.Status != StatusFailed {
				allDone = false
			}

			if task.Status == StatusPending {
				depsSatisfied := true
				mu.Lock()
				for _, dep := range task.Dependencies {
					if !completed[dep] {
						depsSatisfied = false
						break
					}
				}
				mu.Unlock()

				if depsSatisfied {
					task.Status = StatusRunning
					readyTasks = append(readyTasks, task)
				}
			}
		}
		o.mu.Unlock()

		if allDone {
			break
		}

		if len(readyTasks) == 0 {
			time.Sleep(10 * time.Millisecond)
			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				return ctx.Err()
			default:
				continue
			}
		}

		for _, task := range readyTasks {
			wg.Add(1)
			go func(t *Task) {
				defer wg.Done()
				err := t.Execute(ctx)
				o.mu.Lock()
				if err != nil {
					t.Status = StatusFailed
					t.Err = err
					o.mu.Unlock()
					errChan <- err
					cancel()
					return
				}
				t.Status = StatusCompleted
				o.mu.Unlock()

				mu.Lock()
				completed[t.ID] = true
				mu.Unlock()
			}(task)
		}
	}

	wg.Wait()
	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// Saga Test
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

type SagaOrchestrator struct {
	steps []SagaStep
}

func (s *SagaOrchestrator) Execute(ctx context.Context) error {
	executed := make([]SagaStep, 0)
	for _, step := range s.steps {
		if err := step.Execute(ctx); err != nil {
			for i := len(executed) - 1; i >= 0; i-- {
				if executed[i].Compensate != nil {
					_ = executed[i].Compensate(ctx)
				}
			}
			return err
		}
		executed = append(executed, step)
	}
	return nil
}

func TestOrchestrator(t *testing.T) {
	// Test DAG
	dag := NewDAGOrchestrator()
	var step1Done, step2Done, step3Done atomic.Bool

	dag.AddTask("step1", nil, func(ctx context.Context) error {
		step1Done.Store(true)
		return nil
	})
	dag.AddTask("step2", []string{"step1"}, func(ctx context.Context) error {
		if !step1Done.Load() {
			return errors.New("step1 should be done before step2")
		}
		step2Done.Store(true)
		return nil
	})
	dag.AddTask("step3", []string{"step2"}, func(ctx context.Context) error {
		if !step2Done.Load() {
			return errors.New("step2 should be done before step3")
		}
		step3Done.Store(true)
		return nil
	})

	err := dag.Run(context.Background())
	if err != nil {
		t.Fatalf("DAG execution failed: %v", err)
	}
	if !step3Done.Load() {
		t.Fatalf("Step3 did not complete")
	}

	// Test Saga rollback
	var compensatedStep1 atomic.Bool
	saga := &SagaOrchestrator{
		steps: []SagaStep{
			{
				Name:    "charge",
				Execute: func(ctx context.Context) error { return nil },
				Compensate: func(ctx context.Context) error {
					compensatedStep1.Store(true)
					return nil
				},
			},
			{
				Name:    "book",
				Execute: func(ctx context.Context) error { return errors.New("booking failed") },
			},
		},
	}

	err = saga.Execute(context.Background())
	if err == nil {
		t.Fatalf("Expected saga error")
	}
	if !compensatedStep1.Load() {
		t.Fatalf("Expected step1 to be compensated on failure")
	}
}
