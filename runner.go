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

// resolveExecutionOrder computes a valid execution order for all tasks in the workflow.
//
// It ensures that:
// - each task is executed after its dependencies
// - all dependencies exist
// - no cyclic dependencies are present
//
// It returns a slice of tasks sorted in execution order,
// or an error if validation fails.
func resolveExecutionOrder(wf *Workflow) ([]*Task, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	var result []*Task

	var visit func(string) error

	visit = func(name string) error {
		// detect cycle
		if visiting[name] {
			return fmt.Errorf("cycle detected at task: %s", name)
		}

		// already processed
		if visited[name] {
			return nil
		}

		task, exists := wf.Tasks[name]
		if !exists {
			return fmt.Errorf("task not found: %s", name)
		}

		visiting[name] = true

		// visit dependencies first
		for _, dep := range task.DependsOn {
			if _, ok := wf.Tasks[dep]; !ok {
				return fmt.Errorf("task %s depends on unknown task %s", name, dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true

		result = append(result, task)
		return nil
	}

	for name := range wf.Tasks {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// RunWorkflow executes all tasks in the given workflow in dependency order.
//
// It performs the following steps:
// - resolves the correct execution order of tasks based on dependencies
// - prints execution progress using the UI layer
// - executes each task sequentially
// - stops execution on the first failure
//
// Returns an error if:
// - the execution order cannot be resolved (e.g., cycle or missing dependency)
// - any task execution fails
func RunWorkflow(wf *Workflow) error {
	tasks, err := resolveExecutionOrder(wf)
	if err != nil {
		return err
	}

	PrintStart()

	totalStart := time.Now()

	for _, task := range tasks {
		start := time.Now()

		PrintTaskStart(task.Name, task.DependsOn)

		err := runTask(task)

		duration := time.Since(start).Seconds()

		if err != nil {
			PrintTaskFailure(task.Name, duration)
			return err
		}

		PrintTaskSuccess(task.Name, duration)
	}

	PrintEnd(time.Since(totalStart).Seconds())

	return nil
}

// RunWorkflowParallel executes the workflow using the default shell-based executor.
func RunWorkflowParallel(wf *Workflow) error {
	return RunWorkflowParallelWithExecutor(wf, runTask)
}

// RunWorkflowParallelWithExecutor executes the workflow using a custom task executor.
//
// This is primarily used for testing to control task behavior.
func RunWorkflowParallelWithExecutor(wf *Workflow, exec TaskExecutor) error {
	_, err := resolveExecutionOrder(wf)
	if err != nil {
		return err
	}

	PrintStart()
	totalStart := time.Now()

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

	remainingTasks := len(wf.Tasks)

	run := func(task *Task) {
		defer wg.Done()

		start := time.Now()
		PrintTaskStart(task.Name, task.DependsOn)

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

	for name, count := range inDegree {
		if count == 0 {
			wg.Add(1)
			go run(wf.Tasks[name])
		}
	}

	for remainingTasks > 0 {
		select {
		case finished := <-done:
			remainingTasks--

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

	PrintEnd(time.Since(totalStart).Seconds())
	return nil
}