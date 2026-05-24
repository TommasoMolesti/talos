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
	err = validateWorkflowSchema(&root, path)
	if err != nil {
		return nil, err
	}

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

// validateWorkflowSchema rejects unsupported fields before decoding so typos
// cannot silently change workflow behavior.
func validateWorkflowSchema(root *yaml.Node, path string) error {
	var topLevel *yaml.Node = rootMapping(root)
	if topLevel == nil {
		return nil
	}
	if topLevel.Kind != yaml.MappingNode {
		return schemaError(path, nodeLocation(topLevel), "workflow must be a mapping")
	}

	var topLevelFields map[string]bool = map[string]bool{
		"defaults": true,
		"tasks":    true,
	}
	var defaultFields map[string]bool = map[string]bool{
		"cwd":     true,
		"shell":   true,
		"env":     true,
		"retries": true,
		"timeout": true,
	}
	var taskFields map[string]bool = map[string]bool{
		"command":     true,
		"cwd":         true,
		"depends_on":  true,
		"description": true,
		"env":         true,
		"retries":     true,
		"shell":       true,
		"timeout":     true,
	}
	var seenTopLevel map[string]bool = make(map[string]bool)

	for i := 0; i+1 < len(topLevel.Content); i += 2 {
		var key *yaml.Node = topLevel.Content[i]
		var value *yaml.Node = topLevel.Content[i+1]
		if seenTopLevel[key.Value] {
			return schemaError(path, nodeLocation(key), "duplicate top-level field %q", key.Value)
		}
		seenTopLevel[key.Value] = true
		if !topLevelFields[key.Value] {
			return schemaError(path, nodeLocation(key), "unsupported top-level field %q", key.Value)
		}
		if key.Value == "defaults" {
			if value.Kind != yaml.MappingNode {
				return schemaError(path, nodeLocation(value), "defaults must be a mapping")
			}
			var err error = validateMappingFields(path, "defaults", value, defaultFields)
			if err != nil {
				return err
			}
			err = validateDefaultFieldTypes(path, value)
			if err != nil {
				return err
			}
		}
		if key.Value == "tasks" {
			if value.Kind != yaml.MappingNode {
				return schemaError(path, nodeLocation(value), "tasks must be a mapping")
			}
			var err error = validateTaskSchema(path, value, taskFields)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// validateTaskSchema rejects unsupported fields in each task definition.
func validateTaskSchema(path string, tasks *yaml.Node, allowed map[string]bool) error {
	if tasks == nil || tasks.Kind != yaml.MappingNode {
		return nil
	}
	var seenTasks map[string]bool = make(map[string]bool)
	for i := 0; i+1 < len(tasks.Content); i += 2 {
		var taskName *yaml.Node = tasks.Content[i]
		var taskConfig *yaml.Node = tasks.Content[i+1]
		if strings.TrimSpace(taskName.Value) == "" {
			return schemaError(path, nodeLocation(taskName), "task name is required")
		}
		if taskConfig.Kind != yaml.MappingNode {
			return schemaError(path, nodeLocation(taskConfig), "task %q must be a mapping", taskName.Value)
		}
		if seenTasks[taskName.Value] {
			return schemaError(path, nodeLocation(taskName), "duplicate task %q", taskName.Value)
		}
		seenTasks[taskName.Value] = true
		var err error = validateMappingFields(path, fmt.Sprintf("task %q", taskName.Value), taskConfig, allowed)
		if err != nil {
			return err
		}
		err = validateTaskFieldTypes(path, taskName.Value, taskConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

// validateDefaultFieldTypes rejects invalid default field value types before YAML decoding.
func validateDefaultFieldTypes(path string, defaults *yaml.Node) error {
	for i := 0; i+1 < len(defaults.Content); i += 2 {
		var key *yaml.Node = defaults.Content[i]
		var value *yaml.Node = defaults.Content[i+1]
		switch key.Value {
		case "cwd", "shell":
			if !isStringNode(value) {
				return schemaError(path, nodeLocation(value), "%s in defaults must be a string", key.Value)
			}
		case "retries", "timeout":
			if !isIntNode(value) {
				return schemaError(path, nodeLocation(value), "%s in defaults must be an integer", key.Value)
			}
		case "env":
			var err error = validateEnvMapping(path, "defaults", value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// validateTaskFieldTypes rejects invalid task field value types before YAML decoding.
func validateTaskFieldTypes(path string, taskName string, taskConfig *yaml.Node) error {
	for i := 0; i+1 < len(taskConfig.Content); i += 2 {
		var key *yaml.Node = taskConfig.Content[i]
		var value *yaml.Node = taskConfig.Content[i+1]
		switch key.Value {
		case "command", "cwd", "description", "shell":
			if !isStringNode(value) {
				return schemaError(path, nodeLocation(value), "%s in task %q must be a string", key.Value, taskName)
			}
		case "retries", "timeout":
			if !isIntNode(value) {
				return schemaError(path, nodeLocation(value), "%s in task %q must be an integer", key.Value, taskName)
			}
		case "env":
			var err error = validateEnvMapping(path, fmt.Sprintf("task %q", taskName), value)
			if err != nil {
				return err
			}
		case "depends_on":
			var err error = validateDependencyList(path, taskName, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// validateDependencyList rejects invalid dependency lists before YAML decoding.
func validateDependencyList(path string, taskName string, deps *yaml.Node) error {
	if deps.Kind != yaml.SequenceNode {
		return schemaError(path, nodeLocation(deps), "depends_on in task %q must be a list", taskName)
	}
	for _, dep := range deps.Content {
		if !isStringNode(dep) {
			return schemaError(path, nodeLocation(dep), "dependency in task %q must be a string", taskName)
		}
		if strings.TrimSpace(dep.Value) == "" {
			return schemaError(path, nodeLocation(dep), "task %s dependency name is required", taskName)
		}
	}
	return nil
}

// validateEnvMapping rejects invalid env maps before YAML decoding.
func validateEnvMapping(path string, label string, env *yaml.Node) error {
	if env.Kind != yaml.MappingNode {
		return schemaError(path, nodeLocation(env), "env in %s must be a mapping", label)
	}
	for i := 0; i+1 < len(env.Content); i += 2 {
		var key *yaml.Node = env.Content[i]
		var value *yaml.Node = env.Content[i+1]
		if !isStringNode(key) {
			return schemaError(path, nodeLocation(key), "env key in %s must be a string", label)
		}
		if !isStringNode(value) {
			return schemaError(path, nodeLocation(value), "env value for %q in %s must be a string", key.Value, label)
		}
	}
	return nil
}

// isStringNode reports whether a YAML node is an explicit or implicit string.
func isStringNode(node *yaml.Node) bool {
	return node != nil && node.Kind == yaml.ScalarNode && node.ShortTag() == "!!str"
}

// isIntNode reports whether a YAML node is an integer scalar.
func isIntNode(node *yaml.Node) bool {
	return node != nil && node.Kind == yaml.ScalarNode && node.ShortTag() == "!!int"
}

// validateMappingFields rejects unsupported keys from one YAML mapping.
func validateMappingFields(path string, label string, node *yaml.Node, allowed map[string]bool) error {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	var seen map[string]bool = make(map[string]bool)
	for i := 0; i+1 < len(node.Content); i += 2 {
		var key *yaml.Node = node.Content[i]
		if seen[key.Value] {
			return schemaError(path, nodeLocation(key), "duplicate field %q in %s", key.Value, label)
		}
		seen[key.Value] = true
		if !allowed[key.Value] {
			return schemaError(path, nodeLocation(key), "unsupported field %q in %s", key.Value, label)
		}
	}
	return nil
}

// schemaError formats schema validation errors with source locations.
func schemaError(path string, location ConfigLocation, format string, args ...interface{}) error {
	var err error = fmt.Errorf(format, args...)
	if location.Line == 0 {
		return err
	}
	return fmt.Errorf("%s:%d:%d: %w", path, location.Line, location.Column, err)
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
	wf.Defaults.Shell = strings.TrimSpace(wf.Defaults.Shell)

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
	task.Shell = strings.TrimSpace(task.Shell)
	if task.Shell == "" {
		task.Shell = defaults.Shell
	}

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
	for _, field := range []string{"retries", "timeout", "shell", "cwd", "env"} {
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

// validationErrorAtTaskField formats a validation error with a task field location.
func validationErrorAtTaskField(wf *Workflow, taskName string, field string, format string, args ...interface{}) error {
	var err error = fmt.Errorf(format, args...)
	var location ConfigLocation
	if wf != nil && wf.TaskLocations != nil {
		var taskLocations TaskConfigLocations = wf.TaskLocations[taskName]
		if taskLocations.Fields != nil {
			location = taskLocations.Fields[field]
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
	if wf == nil || len(wf.Tasks) == 0 {
		return errors.New("workflow must define at least one task")
	}

	for name, task := range wf.Tasks {
		if task == nil {
			return validationErrorAt(wf, name, "", "task %s is empty", name)
		}
		if strings.TrimSpace(task.Command) == "" {
			return validationErrorAtTaskField(wf, name, "command", "task %s command is required", name)
		}
		var seenDependencies map[string]bool = make(map[string]bool, len(task.DependsOn))
		for _, dep := range task.DependsOn {
			if strings.TrimSpace(dep) == "" {
				return validationErrorAt(wf, name, dep, "task %s dependency name is required", name)
			}
			if seenDependencies[dep] {
				return validationErrorAt(wf, name, dep, "task %s depends on %s more than once", name, dep)
			}
			seenDependencies[dep] = true
		}
	}

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
