package main

import (
	"fmt"
	"os"

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
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	err = yaml.Unmarshal(data, &wf)
	if err != nil {
		return nil, err
	}

	for name, task := range wf.Tasks {
		task.Name = name
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
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

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

		return nil
	}

	for name := range wf.Tasks {
		if err := visit(name); err != nil {
			return err
		}
	}

	return nil
}
