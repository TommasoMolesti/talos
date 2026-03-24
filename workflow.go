package main

import "time"

// Task represents a single unit of work in the workflow.
//
// Each task has:
// - a unique Name (assigned after parsing)
// - a shell Command to execute
// - optional dependencies (DependsOn) referencing other task names
type Task struct {
	Name            string        `yaml:"-"`
	Command         string        `yaml:"command"`
	DependsOn       []string      `yaml:"depends_on"`
	Retries         int           `yaml:"retries"`
	Timeout         string        `yaml:"timeout"`
	TimeoutDuration time.Duration `yaml:"-"`
}

// Workflow represents a collection of tasks forming a DAG (Directed Acyclic Graph).
//
// Tasks are stored in a map where the key is the task name.
// Dependencies between tasks define execution order.
type Workflow struct {
	Tasks map[string]*Task `yaml:"tasks"`
}
