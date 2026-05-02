package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	var root yaml.Node
	err = yaml.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	err = root.Decode(&wf)
	if err != nil {
		return nil, err
	}
	wf.SourcePath = path
	wf.DefaultLocations = collectDefaultLocations(&root)
	wf.TaskLocations = collectTaskLocations(&root)

	var baseDir string = filepath.Dir(path)
	err = normalizeWorkflowDefaults(&wf, baseDir)
	if err != nil {
		return nil, withWorkflowLocation(path, wf, err)
	}

	for name, task := range wf.Tasks {
		task.Name = name
		err = applyTaskDefaults(task, wf.Defaults, baseDir)
		if err != nil {
			return nil, withTaskLocation(path, wf, name, err)
		}
	}

	return &wf, nil
}

// collectDefaultLocations records source locations for workflow default fields.
func collectDefaultLocations(root *yaml.Node) map[string]ConfigLocation {
	var defaults *yaml.Node = mappingValue(rootMapping(root), "defaults")
	if defaults == nil {
		return nil
	}
	return collectFieldLocations(defaults)
}

// collectTaskLocations records source locations for task names, fields, and dependencies.
func collectTaskLocations(root *yaml.Node) map[string]TaskConfigLocations {
	var tasks *yaml.Node = mappingValue(rootMapping(root), "tasks")
	if tasks == nil || tasks.Kind != yaml.MappingNode {
		return nil
	}

	var locations map[string]TaskConfigLocations = make(map[string]TaskConfigLocations)
	for i := 0; i+1 < len(tasks.Content); i += 2 {
		var nameNode *yaml.Node = tasks.Content[i]
		var taskNode *yaml.Node = tasks.Content[i+1]
		var taskLocations TaskConfigLocations = TaskConfigLocations{
			Name:         nodeLocation(nameNode),
			Fields:       collectFieldLocations(taskNode),
			Dependencies: collectDependencyLocations(taskNode),
		}
		locations[nameNode.Value] = taskLocations
	}
	return locations
}

// collectFieldLocations records source locations for mapping field names.
func collectFieldLocations(node *yaml.Node) map[string]ConfigLocation {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	var locations map[string]ConfigLocation = make(map[string]ConfigLocation)
	for i := 0; i+1 < len(node.Content); i += 2 {
		var key *yaml.Node = node.Content[i]
		locations[key.Value] = nodeLocation(key)
	}
	return locations
}

// collectDependencyLocations records source locations for depends_on entries.
func collectDependencyLocations(taskNode *yaml.Node) map[string]ConfigLocation {
	var deps *yaml.Node = mappingValue(taskNode, "depends_on")
	if deps == nil || deps.Kind != yaml.SequenceNode {
		return nil
	}

	var locations map[string]ConfigLocation = make(map[string]ConfigLocation)
	for _, dep := range deps.Content {
		locations[dep.Value] = nodeLocation(dep)
	}
	return locations
}

// rootMapping returns the YAML document's top-level mapping node.
func rootMapping(root *yaml.Node) *yaml.Node {
	if root == nil {
		return nil
	}
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		return root.Content[0]
	}
	return root
}

// mappingValue returns the value node for a mapping key.
func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

// nodeLocation converts a YAML node position into a config location.
func nodeLocation(node *yaml.Node) ConfigLocation {
	if node == nil {
		return ConfigLocation{}
	}
	return ConfigLocation{Line: node.Line, Column: node.Column}
}

// normalizeWorkflowDefaults validates and resolves workflow-level defaults.
func normalizeWorkflowDefaults(wf *Workflow, baseDir string) error {
	if wf.Defaults.RetriesConfig != nil {
		if *wf.Defaults.RetriesConfig < 0 {
			return errors.New("defaults retries must be zero or greater")
		}
		wf.Defaults.Retries = *wf.Defaults.RetriesConfig
	}
	if wf.Defaults.TimeoutConfig != nil {
		if *wf.Defaults.TimeoutConfig < 0 {
			return errors.New("defaults timeout must be zero or greater")
		}
		wf.Defaults.TimeoutSeconds = *wf.Defaults.TimeoutConfig
	}
	if wf.Defaults.Cwd != "" {
		wf.Defaults.WorkingDir = resolveWorkflowPath(baseDir, wf.Defaults.Cwd)
	}
	return nil
}

// applyTaskDefaults merges workflow defaults into one task.
func applyTaskDefaults(task *Task, defaults WorkflowDefaults, baseDir string) error {
	if task.RetriesConfig != nil {
		if *task.RetriesConfig < 0 {
			return errors.New("retries must be zero or greater")
		}
		task.Retries = *task.RetriesConfig
	} else {
		task.Retries = defaults.Retries
	}

	if task.TimeoutConfig != nil {
		if *task.TimeoutConfig < 0 {
			return errors.New("timeout must be zero or greater")
		}
		task.TimeoutSeconds = *task.TimeoutConfig
	} else {
		task.TimeoutSeconds = defaults.TimeoutSeconds
	}
	if task.TimeoutSeconds > 0 {
		task.TimeoutDuration = time.Duration(task.TimeoutSeconds) * time.Second
	}

	task.Env = mergeEnv(defaults.Env, task.Env)

	if task.Cwd != "" {
		task.WorkingDir = resolveWorkflowPath(baseDir, task.Cwd)
		return nil
	}
	task.Cwd = defaults.Cwd
	task.WorkingDir = defaults.WorkingDir
	return nil
}

// resolveWorkflowPath resolves a workflow-relative path to a clean path.
func resolveWorkflowPath(baseDir string, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}

// mergeEnv combines default and task-specific environment variables.
func mergeEnv(defaults map[string]string, overrides map[string]string) map[string]string {
	if len(defaults) == 0 && len(overrides) == 0 {
		return nil
	}

	var merged map[string]string = make(map[string]string, len(defaults)+len(overrides))
	for key, value := range defaults {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	return merged
}

// withWorkflowLocation prefixes an error with its workflow default source location.
func withWorkflowLocation(path string, wf Workflow, err error) error {
	var field string = validationErrorField(err.Error())
	return withLocation(path, ConfigLocation{}, err, func() ConfigLocation {
		if wf.DefaultLocations != nil {
			return wf.DefaultLocations[field]
		}
		return ConfigLocation{}
	}, field)
}

// withTaskLocation prefixes an error with its task source location.
func withTaskLocation(path string, wf Workflow, name string, err error) error {
	var locations TaskConfigLocations = wf.TaskLocations[name]
	var field string = validationErrorField(err.Error())
	return withLocation(path, locations.Name, err, func() ConfigLocation {
		if locations.Fields != nil {
			return locations.Fields[field]
		}
		return ConfigLocation{}
	}, field)
}

// withLocation prefixes an error with file, line, and column when available.
func withLocation(path string, fallback ConfigLocation, err error, locate func() ConfigLocation, field string) error {
	var location ConfigLocation
	if locate != nil && field != "" {
		location = locate()
	}
	if location.Line == 0 {
		location = fallback
	}
	if location.Line == 0 {
		return err
	}
	return fmt.Errorf("%s:%d:%d: %w", path, location.Line, location.Column, err)
}

// validationErrorField infers the config field named by a validation error.
func validationErrorField(message string) string {
	for _, field := range []string{"retries", "timeout", "cwd", "env"} {
		if strings.Contains(message, field) {
			return field
		}
	}
	return ""
}

// validationErrorAt formats a validation error with a task or dependency location.
func validationErrorAt(wf *Workflow, taskName string, depName string, format string, args ...interface{}) error {
	var err error = fmt.Errorf(format, args...)
	var location ConfigLocation
	if wf != nil && wf.TaskLocations != nil {
		var taskLocations TaskConfigLocations = wf.TaskLocations[taskName]
		if depName != "" && taskLocations.Dependencies != nil {
			location = taskLocations.Dependencies[depName]
		}
		if location.Line == 0 {
			location = taskLocations.Name
		}
	}
	if location.Line == 0 {
		return err
	}
	if wf != nil && wf.SourcePath != "" {
		return fmt.Errorf("%s:%d:%d: %w", wf.SourcePath, location.Line, location.Column, err)
	}
	return fmt.Errorf("line %d, column %d: %w", location.Line, location.Column, err)
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
			return validationErrorAt(wf, name, "", "cycle detected at task: %s", name)
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
				return validationErrorAt(wf, name, dep, "task %s depends on unknown task %s", name, dep)
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
