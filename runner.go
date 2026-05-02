package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

// RunOptions defines execution configuration for the workflow.
type RunOptions struct {
	// MaxConcurrency limits the number of tasks running in parallel.
	// If <= 0, no limit is applied.
	MaxConcurrency int
	// DryRun prints the execution plan without running commands.
	DryRun bool
}

type taskResult struct {
	name     string
	err      error
	status   taskStatus
	attempts int
	duration time.Duration
	timeout  time.Duration
}

type taskTimeoutError struct {
	task    string
	timeout time.Duration
}

// Error returns a human-readable timeout failure message.
func (e taskTimeoutError) Error() string {
	return fmt.Sprintf("task %s timed out after %s", e.task, e.timeout)
}

type taskStatus string

const (
	taskStatusPending  taskStatus = "pending"
	taskStatusSuccess  taskStatus = "success"
	taskStatusFailed   taskStatus = "failed"
	taskStatusTimedOut taskStatus = "timed_out"
	taskStatusCanceled taskStatus = "canceled"
	taskStatusSkipped  taskStatus = "skipped"
)

type taskSummary struct {
	Description string
	Status      taskStatus
	Attempts    int
	Duration    time.Duration
	Timeout     time.Duration
}

type executionSummary struct {
	Tasks map[string]*taskSummary
}

// runTask executes a single task command using the system shell.
var runTask func(context.Context, *Task) error = func(ctx context.Context, task *Task) error {
	var cmd *exec.Cmd = exec.CommandContext(ctx, "sh", "-c", task.Command)
	cmd.Dir = taskDir(task)
	cmd.Env = taskEnv(task.Env)

	var output []byte
	var err error
	output, err = cmd.CombinedOutput()

	var lines []string = strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			PrintTaskOutputLine(line)
		}
	}

	return err
}

// taskDir returns the directory a task command should run from.
func taskDir(task *Task) string {
	if task.WorkingDir != "" {
		return task.WorkingDir
	}
	return task.Cwd
}

// taskEnv returns the process environment with task-specific overrides applied.
func taskEnv(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}

	var keys []string = make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var values []string = append([]string(nil), os.Environ()...)
	for _, key := range keys {
		values = append(values, key+"="+env[key])
	}

	return values
}

// RunWorkflowParallel executes the workflow with optional concurrency limits.
func RunWorkflowParallel(wf *Workflow, opts RunOptions) error {
	var err error = validateExecutionOrder(wf)
	if err != nil {
		return err
	}

	if opts.DryRun {
		var plan [][]string
		plan, err = buildExecutionPlan(wf)
		if err != nil {
			return err
		}
		PrintDryRun(plan, wf)
		return nil
	}

	PrintStart()
	var startTotal time.Time = time.Now()
	var summary *executionSummary = newExecutionSummary(wf)

	var inDegree map[string]int = make(map[string]int)
	var dependents map[string][]string = make(map[string][]string)

	for name, task := range wf.Tasks {
		inDegree[name] = len(task.DependsOn)
		for _, dep := range task.DependsOn {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var results chan taskResult = make(chan taskResult, len(wf.Tasks))
	var started int = 0
	var completed int = 0
	var firstErr error

	var sem chan struct{}
	if opts.MaxConcurrency > 0 {
		sem = make(chan struct{}, opts.MaxConcurrency)
	}

	var run func(*Task)
	run = func(task *Task) {
		defer wg.Done()

		if sem != nil {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results <- taskResult{name: task.Name, err: ctx.Err(), status: taskStatusSkipped}
				return
			}
		}

		if ctx.Err() != nil {
			results <- taskResult{name: task.Name, err: ctx.Err(), status: taskStatusSkipped}
			return
		}

		var taskCtx context.Context = ctx
		var cancelTask context.CancelFunc
		if task.TimeoutDuration > 0 {
			taskCtx, cancelTask = context.WithTimeout(ctx, task.TimeoutDuration)
			defer cancelTask()
		}

		var start time.Time = time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

		var maxAttempts int = task.Retries + 1
		var err error
		var attempts int = 0
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			attempts = attempt
			err = runTask(taskCtx, task)
			if err == nil {
				break
			}
			if task.TimeoutDuration > 0 && errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
				break
			}
			if errors.Is(err, context.Canceled) || attempt == maxAttempts {
				break
			}
			PrintTaskRetry(task.Name, attempt+1, maxAttempts, err)
		}

		var duration time.Duration = time.Since(start)

		if err != nil {
			if task.TimeoutDuration > 0 && errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
				PrintTaskTimeout(task.Name, duration.Seconds(), task.TimeoutDuration)
				results <- taskResult{
					name:     task.Name,
					err:      taskTimeoutError{task: task.Name, timeout: task.TimeoutDuration},
					status:   taskStatusTimedOut,
					attempts: attempts,
					duration: duration,
					timeout:  task.TimeoutDuration,
				}
				return
			}
			if errors.Is(err, context.Canceled) {
				PrintTaskCanceled(task.Name, duration.Seconds())
				results <- taskResult{
					name:     task.Name,
					err:      err,
					status:   taskStatusCanceled,
					attempts: attempts,
					duration: duration,
				}
			} else {
				PrintTaskFailure(task.Name, duration.Seconds())
				results <- taskResult{
					name:     task.Name,
					err:      err,
					status:   taskStatusFailed,
					attempts: attempts,
					duration: duration,
				}
			}
			return
		}

		PrintTaskSuccess(task.Name, duration.Seconds())
		results <- taskResult{
			name:     task.Name,
			status:   taskStatusSuccess,
			attempts: attempts,
			duration: duration,
		}
	}

	var startTask func(string)
	startTask = func(name string) {
		if ctx.Err() != nil {
			return
		}
		wg.Add(1)
		started++
		go run(wf.Tasks[name])
	}

	for name, count := range inDegree {
		if count == 0 {
			startTask(name)
		}
	}

	for completed < started {
		var result taskResult = <-results
		completed++
		summary.record(result)

		if result.err != nil {
			if firstErr == nil && !errors.Is(result.err, context.Canceled) {
				firstErr = result.err
				cancel()
			}
			continue
		}

		for _, dep := range dependents[result.name] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				startTask(dep)
			}
		}
	}

	wg.Wait()
	summary.markPendingAsSkipped()
	var totalDuration time.Duration = time.Since(startTotal)
	PrintSummary(summary)
	if firstErr != nil {
		PrintEnd(totalDuration.Seconds(), false)
		return firstErr
	}
	PrintEnd(totalDuration.Seconds(), true)

	return nil
}

// buildExecutionPlan groups tasks into deterministic execution stages for dry-runs.
//
// Each stage contains all tasks whose dependencies have already been satisfied by
// earlier stages, which means the tasks in the same inner slice can run in
// parallel. Task names are sorted within a stage so dry-run output remains stable
// across map iteration order. The function returns an error if not all tasks can
// be scheduled into the plan.
func buildExecutionPlan(wf *Workflow) ([][]string, error) {
	var inDegree map[string]int = make(map[string]int, len(wf.Tasks))
	var dependents map[string][]string = make(map[string][]string, len(wf.Tasks))

	for name, task := range wf.Tasks {
		inDegree[name] = len(task.DependsOn)
		for _, dep := range task.DependsOn {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	var ready []string
	for name, count := range inDegree {
		if count == 0 {
			ready = append(ready, name)
		}
	}
	sort.Strings(ready)

	var plan [][]string
	var processed int = 0

	for len(ready) > 0 {
		var stage []string = append([]string(nil), ready...)
		plan = append(plan, stage)
		processed += len(stage)

		var next []string
		for _, name := range stage {
			for _, dep := range dependents[name] {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					next = append(next, dep)
				}
			}
		}
		sort.Strings(next)
		ready = next
	}

	if processed != len(wf.Tasks) {
		return nil, errors.New("unable to build execution plan")
	}

	return plan, nil
}

// newExecutionSummary initializes pending summary entries for every task.
func newExecutionSummary(wf *Workflow) *executionSummary {
	var tasks map[string]*taskSummary = make(map[string]*taskSummary, len(wf.Tasks))
	for name, task := range wf.Tasks {
		tasks[name] = &taskSummary{
			Description: task.Description,
			Status:      taskStatusPending,
		}
	}
	return &executionSummary{Tasks: tasks}
}

// record stores the final result for one task in the execution summary.
func (s *executionSummary) record(result taskResult) {
	var task *taskSummary = s.Tasks[result.name]
	if task == nil {
		return
	}
	task.Status = result.status
	task.Attempts = result.attempts
	task.Duration = result.duration
	task.Timeout = result.timeout
}

// markPendingAsSkipped marks never-started tasks as skipped after fail-fast cancellation.
func (s *executionSummary) markPendingAsSkipped() {
	for _, task := range s.Tasks {
		if task.Status == taskStatusPending {
			task.Status = taskStatusSkipped
		}
	}
}
