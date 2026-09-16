---
id: orchestrator_in_go
aliases:
  - Orchestrator in Go
  - Building an Orchestrator in Go
  - Distributed Workflow Orchestration
tags:
  - go
  - orchestration
  - distributed-systems
  - architecture
  - saga-pattern
  - kubernetes
  - workflows
dg-publish: true
---

# Building an Orchestrator in Go

An **Orchestrator** is a centralized coordination system responsible for scheduling, dispatching, managing dependencies, monitoring health, and guaranteeing fault tolerance across distributed tasks or services.

Go is the foundational language of the modern orchestration ecosystem: **Kubernetes, Docker, Nomad, Temporal Server, and Argo Workflows are all written in Go**.

```
                           +-------------------------------------+
                           |          Client / Trigger           |
                           +-------------------------------------+
                                              │
                                              ▼
+───────────────────────────────────────────────────────────────────────────────────────────────+
|                                    ORCHESTRATOR ENGINE (Go)                                   |
|                                                                                               |
|  +--------------------+     +----------------------+     +---------------------------------+  |
|  |  DAG Planner       | ──> | State Machine / DB   | ──> | Worker Dispatcher (Concurrency) |  |
|  |  (Dependency Graph)|     | (Pending -> Running) |     | (Channels, Rate-limits, Pools)  |  |
|  +--------------------+     +----------------------+     +---------------------------------+  |
|                                                                          │                    |
|  +--------------------+     +----------------------+                     │                    |
|  | Reconciliation Loop| <── | Fault Tolerance      | <───────────────────┘                    |
|  | (Desired vs Actual)|     | (Retries, Sagas/Roll)|                                          |
|  +--------------------+     +----------------------+                                          |
+───────────────────────────────────────────────────────────────────────────────────────────────+
                                              │
                     ┌────────────────────────┼────────────────────────┐
                     ▼                        ▼                        ▼
              [ Payment API ]          [ Inventory API ]         [ Shipping API ]
```

---

## 1. Taxonomy of Orchestrators

| Orchestrator Category | Core Responsibility | State Model | Real-World Examples |
|---|---|---|---|
| **Workflow / DAG Engine** | Executes steps according to dependency graphs | Persistent Event History / Checkpoints | **Temporal, Airflow, Argo Workflows** |
| **Saga Orchestrator** | Distributed transactions with automatic rollbacks | Finite State Machine (Forward / Compensate) | **Camunda, Uber Cadence, AWS Step Functions** |
| **Container / Resource** | Assigns tasks to cluster nodes to match desired state | Declarative Reconciliation Loop | **Kubernetes, HashiCorp Nomad, Docker Swarm** |
| **Data / Pipeline Engine** | High-throughput batch or stream data processing | Directed Acyclic Graph (DAG) | **Apache Spark, Prefect, Dagster** |

---

## 2. Architecture 1: The DAG Workflow Orchestrator in Go

In a DAG workflow, tasks can only execute after all their parent dependencies complete successfully. Independent branches can run concurrently in parallel goroutines.

```
       [ Step A: Download Raw Data ]
             /               \
            ▼                 ▼
   [ Step B: Clean Text ]   [ Step C: Extract Audio ]
            \                 /
             ▼               ▼
        [ Step D: Combine & Index ]
```

### Complete Implementation

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "PENDING"
	StatusRunning   TaskStatus = "RUNNING"
	StatusCompleted TaskStatus = "COMPLETED"
	StatusFailed    TaskStatus = "FAILED"
)

// TaskFn represents executable work
type TaskFn func(ctx context.Context) error

// Task defines a unit of execution within the DAG
type Task struct {
	ID           string
	Execute      TaskFn
	Dependencies []string // Task IDs that must finish before this runs
	Status       TaskStatus
	Err          error
}

// DAGOrchestrator manages and executes task graphs concurrently
type DAGOrchestrator struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

func NewDAGOrchestrator() *DAGOrchestrator {
	return &DAGOrchestrator{
		tasks: make(map[string]*Task),
	}
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

// Run executes the DAG respecting dependencies, running independent tasks in parallel
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
				// Check if all dependencies are completed
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
			// No tasks ready, verify if tasks are running or deadlock exists
			time.Sleep(50 * time.Millisecond)
			select {
			case err := <-errChan:
				return err
			case <-ctx.Done():
				return ctx.Err()
			default:
				continue
			}
		}

		// Launch all ready tasks in parallel
		for _, task := range readyTasks {
			wg.Add(1)
			go func(t *Task) {
				defer wg.Done()
				fmt.Printf("[Orchestrator] Starting task: %s\n", t.ID)

				err := t.Execute(ctx)

				o.mu.Lock()
				if err != nil {
					t.Status = StatusFailed
					t.Err = err
					o.mu.Unlock()
					errChan <- fmt.Errorf("task %s failed: %w", t.ID, err)
					cancel() // Abort entire DAG on failure
					return
				}

				t.Status = StatusCompleted
				o.mu.Unlock()

				mu.Lock()
				completed[t.ID] = true
				mu.Unlock()

				fmt.Printf("[Orchestrator] Finished task: %s\n", t.ID)
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
```

---

## 3. Architecture 2: The Distributed Saga Orchestrator

In microservice architectures, distributed transactions across multiple databases cannot rely on 2-Phase Commit (2PC).
The **Saga Pattern** breaks transactions into steps. If step $N$ fails, the orchestrator executes **compensating transactions** in reverse order ($N-1, \dots, 1$) to undo previous side-effects.

```
Forward Execution:
  [ 1. Reserve Inventory ] ──> [ 2. Charge Credit Card ] ──> [ 3. Book Flight (FAILS!) ]
                                                                      │
Compensating Rollback:                                                ▼
  [ 1. Release Inventory ] <── [ 2. Refund Credit Card ] <────────────┘
```

### Complete Saga Implementation

```go
package main

import (
	"context"
	"fmt"
)

// SagaStep defines forward action and its reverse compensation action
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

type SagaOrchestrator struct {
	steps []SagaStep
}

func NewSaga() *SagaOrchestrator {
	return &SagaOrchestrator{steps: make([]SagaStep, 0)}
}

func (s *SagaOrchestrator) AddStep(name string, exec, comp func(ctx context.Context) error) {
	s.steps = append(s.steps, SagaStep{
		Name:       name,
		Execute:    exec,
		Compensate: comp,
	})
}

// Execute runs all steps. If any step fails, compensation runs in reverse.
func (s *SagaOrchestrator) Execute(ctx context.Context) error {
	executedSteps := make([]SagaStep, 0)

	for _, step := range s.steps {
		fmt.Printf("[Saga Forward] Executing step: %s\n", step.Name)
		err := step.Execute(ctx)
		if err != nil {
			fmt.Printf("[Saga ERROR] Step '%s' failed: %v. Initiating rollback!\n", step.Name, err)
			s.rollback(ctx, executedSteps)
			return fmt.Errorf("saga aborted at step %s: %w", step.Name, err)
		}
		executedSteps = append(executedSteps, step)
	}

	fmt.Println("[Saga SUCCESS] All steps executed cleanly.")
	return nil
}

// rollback unwinds previously committed steps in reverse order
func (s *SagaOrchestrator) rollback(ctx context.Context, executed []SagaStep) {
	for i := len(executed) - 1; i >= 0; i-- {
		step := executed[i]
		if step.Compensate != nil {
			fmt.Printf("[Saga Rollback] Compensating step: %s\n", step.Name)
			if err := step.Compensate(ctx); err != nil {
				// In production: send to Dead Letter Queue (DLQ) or alert ops
				fmt.Printf("[CRITICAL ALERT] Compensation failed for %s: %v\n", step.Name, err)
			}
		}
	}
}

func ExampleSagaUsage() {
	ctx := context.Background()
	saga := NewSaga()

	// Step 1: Inventory
	saga.AddStep("ReserveInventory",
		func(ctx context.Context) error {
			fmt.Println("  -> Deducted item quantity from warehouse DB")
			return nil
		},
		func(ctx context.Context) error {
			fmt.Println("  <- Restored item quantity to warehouse DB")
			return nil
		},
	)

	// Step 2: Payment
	saga.AddStep("ProcessPayment",
		func(ctx context.Context) error {
			fmt.Println("  -> Charged $250 to customer card via Stripe")
			return nil
		},
		func(ctx context.Context) error {
			fmt.Println("  <- Issued refund of $250 via Stripe")
			return nil
		},
	)

	// Step 3: Shipping (simulating failure)
	saga.AddStep("DispatchCourier",
		func(ctx context.Context) error {
			return errors.New("courier API timeout / delivery unavailable")
		},
		func(ctx context.Context) error {
			fmt.Println("  <- Cancelled courier dispatch")
			return nil
		},
	)

	_ = saga.Execute(ctx)
}
```

---

## 4. Architecture 3: The Kubernetes Reconciliation Loop

In container and resource orchestrators (Kubernetes, Nomad), the core engine does not just run tasks once. It runs an **infinite control loop** that constantly compares the **Desired State** with the **Actual State** and drives convergence.

$$\text{Reconcile}() : \Delta = \text{DesiredState} - \text{ActualState} \implies \text{Apply}(\Delta)$$

```
     +-----------------+
     |  Desired State  | (e.g., Replicas = 3)
     +-----------------+
              │
              ▼
     [ Reconciler Loop ] <────── [ Observe Cluster State ] (Actual = 2)
              │
      Diff = +1 Pod Needed
              │
              ▼
     [ Create New Pod ] ────────> Cluster converges to 3
```

```go
package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type WorkloadSpec struct {
	Name            string
	DesiredReplicas int
}

type Pod struct {
	ID   string
	Name string
}

// Controller continuously reconciles desired vs actual state
type Controller struct {
	spec       WorkloadSpec
	activePods map[string]*Pod
	mu         sync.Mutex
}

func NewController(spec WorkloadSpec) *Controller {
	return &Controller{
		spec:       spec,
		activePods: make(map[string]*Pod),
	}
}

// Reconcile calculates drift and corrects it
func (c *Controller) Reconcile(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	actualCount := len(c.activePods)
	desiredCount := c.spec.DesiredReplicas

	fmt.Printf("[Reconciler] '%s' -> Desired: %d | Actual: %d\n", c.spec.Name, desiredCount, actualCount)

	if actualCount < desiredCount {
		diff := desiredCount - actualCount
		fmt.Printf("  ==> Scaling UP: creating %d pod(s)\n", diff)
		for i := 0; i < diff; i++ {
			podID := fmt.Sprintf("%s-pod-%d", c.spec.Name, time.Now().UnixNano()%10000)
			c.activePods[podID] = &Pod{ID: podID, Name: c.spec.Name}
		}
	} else if actualCount > desiredCount {
		diff := actualCount - desiredCount
		fmt.Printf("  ==> Scaling DOWN: terminating %d pod(s)\n", diff)
		for podID := range c.activePods {
			if diff <= 0 {
				break
			}
			delete(c.activePods, podID)
			diff--
		}
	} else {
		fmt.Println("  ==> State is steady. No changes required.")
	}

	return nil
}

func (c *Controller) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = c.Reconcile(ctx)
		}
	}
}
```

---

## 5. Production Deep-Dive: Real-World Complex Orchestrators

### 1. Temporal / Cadence (Event Sourcing & Replay Engine)
- **Concept**: *Durable Execution*. Your Go code looks like simple, sequential imperative code, but can pause for 30 days, survive complete machine reboots, and resume exactly where it stopped.
- **How It Works**:
  - The orchestrator maintains an **append-only event log** (e.g. `WorkflowTaskScheduled`, `ActivityTaskCompleted`).
  - When a worker dies, a new worker reloads the history and **replays** the code deterministically to reconstruct the exact stack frame.
- **Industry Use Cases**:
  - Subscription billing cycles (30-day sleep with renewal webhooks).
  - Multi-day financial transactions (ACH, wire clearing).
  - Infrastructure provisioning (Terraform / Cloud formation state coordinator).

### 2. Kubernetes / Nomad (Container & Resource Schedulers)
- **Concept**: Two-Level Scheduler with Raft Consensus.
  - **etcd**: Consistent, distributed key-value store for desired state.
  - **kube-scheduler**: Filters nodes (predicates/affinity) and scores them (bin-packing CPU/memory).
  - **kubelet**: Node-level agent executing containers via CRI (Containerd).
- **Industry Use Cases**:
  - Global container cluster management.
  - Self-healing batch workloads.

### 3. Argo Workflows & Airflow (Data & ML Pipelines)
- **Concept**: Cloud-native DAG engine where each step is an isolated container.
- **Industry Use Cases**:
  - Distributed machine learning model training (data extraction -> embedding -> fine-tuning -> evaluation).
  - Daily ETL/ELT pipelines across petabyte data warehouses.

---

## 6. Orchestration vs. Choreography: Architectural Decision Matrix

| Dimension | Centralized Orchestration | Event Choreography (Event-Driven) |
|---|---|---|
| **Coordination** | Central brain controls execution | Services listen to events and decide actions |
| **Visibility** | Complete end-to-end trace in one dashboard | Hard to trace across distributed pub/sub topics |
| **Coupling** | Services coupled to the orchestrator | Loose coupling; producers don't know consumers |
| **Error Handling** | Clean rollback / Saga management | Difficult; cascading rollback events required |
| **Best Used For** | Complex multi-step transactions (order checkout, cloud provisioning) | Independent side effects (send analytics, welcome email, audit log) |
