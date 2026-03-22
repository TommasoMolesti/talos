package main

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// RunOptions defines execution configuration for the workflow.
type RunOptions struct {
	// MaxConcurrency limits the number of tasks running in parallel.
	// If <= 0, no limit is applied.
	MaxConcurrency int
}

type taskResult struct {
	name string
	err  error
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

		start := time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

		err := runTask(ctx, task)

		duration := time.Since(start).Seconds()

		if err != nil {
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
