package main

import (
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

// runTask executes a single task command using the system shell.
var runTask = func(task *Task) error {
	cmd := exec.Command("sh", "-c", task.Command)

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

	var wg sync.WaitGroup
	done := make(chan string)
	errChan := make(chan error, 1)

	remaining := len(wf.Tasks)

	var sem chan struct{}
	if opts.MaxConcurrency > 0 {
		sem = make(chan struct{}, opts.MaxConcurrency)
	}

	run := func(task *Task) {
		defer wg.Done()

		if sem != nil {
			sem <- struct{}{}
			defer func() { <-sem }()
		}

		start := time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

		err := runTask(task)

		duration := time.Since(start).Seconds()

		if err != nil {
			PrintTaskFailure(task.Name, duration)
			select {
			case errChan <- err:
			default:
			}
			return
		}

		PrintTaskSuccess(task.Name, duration)
		done <- task.Name
	}

	for name, count := range inDegree {
		if count == 0 {
			wg.Add(1)
			go run(wf.Tasks[name])
		}
	}

	for remaining > 0 {
		select {
		case finished := <-done:
			remaining--

			for _, dep := range dependents[finished] {
				inDegree[dep]--
				if inDegree[dep] == 0 {
					wg.Add(1)
					go run(wf.Tasks[dep])
				}
			}

		case err := <-errChan:
			wg.Wait()
			return err
		}
	}

	wg.Wait()
	PrintEnd(time.Since(startTotal).Seconds())

	return nil
}