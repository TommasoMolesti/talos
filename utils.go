package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// loadWorkflow reads a YAML file from the given path and parses it into a Workflow.
//
// It performs the following steps:
// - reads the file content from disk
// - unmarshals the YAML into a Workflow struct
// - assigns task names based on the map keys
//
// Returns the populated Workflow or an error if:
// - the file cannot be read
// - the YAML is invalid or cannot be parsed
func loadWorkflow(path string) (*Workflow, error) {
	var data []byte
	var err error
	data, err = os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	err = yaml.Unmarshal(data, &wf)
	if err != nil {
		return nil, err
	}

	var baseDir string = filepath.Dir(path)
	for name, task := range wf.Tasks {
		task.Name = name
		if task.Retries < 0 {
			return nil, fmt.Errorf("task %s retries must be zero or greater", name)
		}
		if task.TimeoutSeconds < 0 {
			return nil, fmt.Errorf("task %s timeout must be zero or greater", name)
		}
		if task.TimeoutSeconds > 0 {
			task.TimeoutDuration = time.Duration(task.TimeoutSeconds) * time.Second
		}
		if task.Cwd != "" {
			if filepath.IsAbs(task.Cwd) {
				task.WorkingDir = filepath.Clean(task.Cwd)
			} else {
				task.WorkingDir = filepath.Clean(filepath.Join(baseDir, task.Cwd))
			}
		}
	}

	return &wf, nil
}

// validateExecutionOrder computes a valid execution order for all tasks in the workflow.
//
// It ensures that:
// - each task is executed after its dependencies
// - all dependencies exist
// - no cyclic dependencies are present
//
// It returns an error if validation fails.
func validateExecutionOrder(wf *Workflow) error {
	var visited map[string]bool = make(map[string]bool)
	var visiting map[string]bool = make(map[string]bool)

	var visit func(string) error

	visit = func(name string) error {
		var err error
		// detect cycle
		if visiting[name] {
			return fmt.Errorf("cycle detected at task: %s", name)
		}

		// already processed
		if visited[name] {
			return nil
		}

		var task *Task
		var exists bool
		task, exists = wf.Tasks[name]
		if !exists {
			return fmt.Errorf("task not found: %s", name)
		}

		visiting[name] = true

		// visit dependencies first
		for _, dep := range task.DependsOn {
			var ok bool
			_, ok = wf.Tasks[dep]
			if !ok {
				return fmt.Errorf("task %s depends on unknown task %s", name, dep)
			}
			err = visit(dep)
			if err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true

		return nil
	}

	for name := range wf.Tasks {
		var err error = visit(name)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateWorkflow performs config-only checks before execution.
func validateWorkflow(wf *Workflow) error {
	return validateExecutionOrder(wf)
}

// workflowForTarget returns a workflow containing only the target task and the
// dependencies required to execute it.
func workflowForTarget(wf *Workflow, target string) (*Workflow, error) {
	var ok bool
	_, ok = wf.Tasks[target]
	if !ok {
		return nil, fmt.Errorf("target task %s not found", target)
	}

	var included map[string]bool = make(map[string]bool)
	var include func(string) error

	include = func(name string) error {
		var task *Task
		task, ok = wf.Tasks[name]
		if !ok {
			return fmt.Errorf("task not found: %s", name)
		}
		if included[name] {
			return nil
		}
		included[name] = true
		for _, dep := range task.DependsOn {
			_, ok = wf.Tasks[dep]
			if !ok {
				return fmt.Errorf("task %s depends on unknown task %s", name, dep)
			}
			var err error = include(dep)
			if err != nil {
				return err
			}
		}
		return nil
	}

	var err error = include(target)
	if err != nil {
		return nil, err
	}

	var filtered map[string]*Task = make(map[string]*Task, len(included))
	for name := range included {
		filtered[name] = wf.Tasks[name]
	}

	return &Workflow{Tasks: filtered}, nil
}
