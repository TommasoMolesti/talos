package main

import "time"

// Task represents a single unit of work in the workflow.
//
// Each task has:
// - a unique Name (assigned after parsing)
// - a shell Command to execute
// - optional dependencies (DependsOn) referencing other task names
type Task struct {
	Name            string            `yaml:"-"`
	Description     string            `yaml:"description"`
	Command         string            `yaml:"command"`
	Cwd             string            `yaml:"cwd"`
	WorkingDir      string            `yaml:"-"`
	Env             map[string]string `yaml:"env"`
	DependsOn       []string          `yaml:"depends_on"`
	Retries         int               `yaml:"-"`
	RetriesConfig   *int              `yaml:"retries"`
	TimeoutSeconds  int               `yaml:"-"`
	TimeoutConfig   *int              `yaml:"timeout"`
	TimeoutDuration time.Duration     `yaml:"-"`
}

// WorkflowDefaults defines values inherited by tasks that do not override them.
type WorkflowDefaults struct {
	Cwd            string            `yaml:"cwd"`
	WorkingDir     string            `yaml:"-"`
	Env            map[string]string `yaml:"env"`
	Retries        int               `yaml:"-"`
	RetriesConfig  *int              `yaml:"retries"`
	TimeoutSeconds int               `yaml:"-"`
	TimeoutConfig  *int              `yaml:"timeout"`
}

// Workflow represents a collection of tasks forming a DAG (Directed Acyclic Graph).
//
// Tasks are stored in a map where the key is the task name.
// Dependencies between tasks define execution order.
type Workflow struct {
	Defaults WorkflowDefaults `yaml:"defaults"`
	Tasks    map[string]*Task `yaml:"tasks"`
}
