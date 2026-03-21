package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"sync"
)

// TaskExecutor defines a function that executes a task.
//
// This abstraction allows injecting custom behavior during testing.
type TaskExecutor func(task *Task) error

// runTask executes a single task command using the system shell.
//
// It captures both stdout and stderr, prints each output line
// using the UI layer, and returns an error if the command fails.
func runTask(task *Task) error {
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

// RunWorkflowParallel executes the workflow using the default shell-based executor.
func RunWorkflowParallel(wf *Workflow) error {
	return RunWorkflowParallelWithExecutor(wf, runTask)
}

// RunWorkflowParallelWithExecutor executes the workflow using a custom task executor.
//
// It schedules tasks based on their dependencies (DAG) and runs independent tasks in parallel.
// The execution is event-driven: when a task completes, its dependents may become runnable.
//
// The function stops on the first error and returns it.
func RunWorkflowParallelWithExecutor(wf *Workflow, exec TaskExecutor) error {
	err := resolveExecutionOrder(wf)
	if err != nil {
		return err
	}

	PrintStart()
	totalStart := time.Now()

	// inDegree tracks how many dependencies each task still has
	inDegree := make(map[string]int)

	// dependents maps a task to the list of tasks that depend on it
	dependents := make(map[string][]string)

	for name, task := range wf.Tasks {
		inDegree[name] = len(task.DependsOn)

		for _, dep := range task.DependsOn {
			dependents[dep] = append(dependents[dep], name)
		}
	}

	// WaitGroup tracks running goroutines (tasks)
	var wg sync.WaitGroup

	// done channel is used by tasks to signal completion (send task name)
	done := make(chan string)

	errChan := make(chan error, 1)

	// remainingTasks tracks how many tasks are left to complete
	remainingTasks := len(wf.Tasks)

	// run is the worker function that executes a single task
	run := func(task *Task) {
		defer wg.Done()

		start := time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

		// Execute the task using the provided executor
		err := exec(task)

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

	// Start all tasks that have no dependencies (inDegree == 0)
	for name, count := range inDegree {
		if count == 0 {
			wg.Add(1)
			go run(wf.Tasks[name])
		}
	}

	// Main scheduler loop:
	// waits for tasks to finish and schedules new ones when ready
	for remainingTasks > 0 {
		select {
		case finished := <-done:
			remainingTasks--

			// For each task that depends on the finished one
			for _, dep := range dependents[finished] {
				inDegree[dep]--

				// If all dependencies are satisfied, schedule the task
				if inDegree[dep] == 0 {
					wg.Add(1)
					go run(wf.Tasks[dep])
				}
			}

		// An error occurred in one of the tasks
		case err := <-errChan:
			// Wait for already running tasks to finish before exiting
			wg.Wait()
			return err
		}
	}

	// Wait for all running goroutines to complete
	wg.Wait()

	PrintEnd(time.Since(totalStart).Seconds())

	return nil
}