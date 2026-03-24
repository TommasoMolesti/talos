package main

import (
	"context"
	"errors"
	"fmt"
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
	name string
	err  error
}

type taskTimeoutError struct {
	task    string
	timeout time.Duration
}

func (e taskTimeoutError) Error() string {
	return fmt.Sprintf("task %s timed out after %s", e.task, e.timeout)
}

// runTask executes a single task command using the system shell.
var runTask = func(ctx context.Context, task *Task) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", task.Command)

	output, err := cmd.CombinedOutput()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			PrintTaskOutputLine(line)
		}
	}

	return err
}

// RunWorkflowParallel executes the workflow with optional concurrency limits.
func RunWorkflowParallel(wf *Workflow, opts RunOptions) error {
	if err := validateExecutionOrder(wf); err != nil {
		return err
	}

	if opts.DryRun {
		plan, err := buildExecutionPlan(wf)
		if err != nil {
			return err
		}
		PrintDryRun(plan, wf)
		return nil
	}

	PrintStart()
	startTotal := time.Now()

	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	for name, task := range wf.Tasks {
		inDegree[name] = len(task.DependsOn)
		for _, dep := range task.DependsOn {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	results := make(chan taskResult, len(wf.Tasks))
	started := 0
	completed := 0
	var firstErr error

	var sem chan struct{}
	if opts.MaxConcurrency > 0 {
		sem = make(chan struct{}, opts.MaxConcurrency)
	}

	run := func(task *Task) {
		defer wg.Done()

		if sem != nil {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results <- taskResult{name: task.Name, err: ctx.Err()}
				return
			}
		}

		if ctx.Err() != nil {
			results <- taskResult{name: task.Name, err: ctx.Err()}
			return
		}

		taskCtx := ctx
		var cancelTask context.CancelFunc
		if task.TimeoutDuration > 0 {
			taskCtx, cancelTask = context.WithTimeout(ctx, task.TimeoutDuration)
			defer cancelTask()
		}

		start := time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

		err := runTask(taskCtx, task)

		duration := time.Since(start).Seconds()

		if err != nil {
			if task.TimeoutDuration > 0 && errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
				PrintTaskTimeout(task.Name, duration, task.TimeoutDuration)
				results <- taskResult{name: task.Name, err: taskTimeoutError{task: task.Name, timeout: task.TimeoutDuration}}
				return
			}
			if errors.Is(err, context.Canceled) {
				PrintTaskCanceled(task.Name, duration)
			} else {
				PrintTaskFailure(task.Name, duration)
			}
			results <- taskResult{name: task.Name, err: err}
			return
		}

		PrintTaskSuccess(task.Name, duration)
		results <- taskResult{name: task.Name}
	}

	startTask := func(name string) {
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
		result := <-results
		completed++

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
	if firstErr != nil {
		return firstErr
	}
	PrintEnd(time.Since(startTotal).Seconds())

	return nil
}

func buildExecutionPlan(wf *Workflow) ([][]string, error) {
	inDegree := make(map[string]int, len(wf.Tasks))
	dependents := make(map[string][]string, len(wf.Tasks))

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
	processed := 0

	for len(ready) > 0 {
		stage := append([]string(nil), ready...)
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
