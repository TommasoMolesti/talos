package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

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