package main

import (
	"testing"
	"time"
	"sync"
)

// TestValidateExecutionOrder_Linear verifies that a simple linear dependency chain
// is validated in the correct order.
func TestValidateExecutionOrder_Linear(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"B"}},
		},
	}

	err := validateExecutionOrder(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestValidateExecutionOrder_MissingDependency verifies that an error is returned
// when a task depends on a non-existent task.
func TestResolveExecutionOrder_MissingDependency(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
		},
	}

	err := validateExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

// TestValidateExecutionOrder_Cycle verifies that cyclic dependencies are detected.
func TestResolveExecutionOrder_Cycle(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
			"B": {Name: "B", DependsOn: []string{"A"}},
		},
	}

	err := validateExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for cycle, got nil")
	}
}

// TestRunWorkflowParallel_AllTasksExecuted verifies that all tasks are executed.
func TestRunWorkflowParallel_AllTasksExecuted(t *testing.T) {
	var mu sync.Mutex
	executed := make(map[string]bool)

	exec := func(task *Task) error {
		mu.Lock()
		executed[task.Name] = true
		mu.Unlock()
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"A"}},
			"D": {Name: "D", DependsOn: []string{"B", "C"}},
		},
	}

	err := RunWorkflowParallelWithExecutor(wf, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name := range wf.Tasks {
		if !executed[name] {
			t.Fatalf("task %s was not executed", name)
		}
	}
}

// TestRunWorkflowParallel_RespectsDependencies verifies that tasks are not executed
// before their dependencies are completed.
func TestRunWorkflowParallel_RespectsDependencies(t *testing.T) {
	var mu sync.Mutex
	completed := make(map[string]bool)

	exec := func(task *Task) error {
		mu.Lock()
		for _, dep := range task.DependsOn {
			if !completed[dep] {
				t.Fatalf("task %s ran before dependency %s", task.Name, dep)
			}
		}
		completed[task.Name] = true
		mu.Unlock()

		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"A"}},
		},
	}

	err := RunWorkflowParallelWithExecutor(wf, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunWorkflowParallel_Parallelism verifies that tasks run concurrently.
func TestRunWorkflowParallel_Parallelism(t *testing.T) {
	exec := func(task *Task) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B"},
			"C": {Name: "C"},
		},
	}

	start := time.Now()

	err := RunWorkflowParallelWithExecutor(wf, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	duration := time.Since(start)

	// If sequential, this would be ~300ms
	if duration > 250*time.Millisecond {
		t.Fatalf("expected parallel execution, took too long: %v", duration)
	}
}

// TestRunWorkflowParallel_Failure verifies that execution stops on error.
func TestRunWorkflowParallel_Failure(t *testing.T) {
	exec := func(task *Task) error {
		if task.Name == "B" {
			return assertError{}
		}
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
		},
	}

	err := RunWorkflowParallelWithExecutor(wf, exec)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestRunWorkflowParallel_DAG verifies correct execution of a multi-level DAG.
func TestRunWorkflowParallel_DAG(t *testing.T) {
	var mu sync.Mutex
	order := []string{}

	exec := func(task *Task) error {
		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		order = append(order, task.Name)
		mu.Unlock()

		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"A"}},
			"D": {Name: "D", DependsOn: []string{"B", "C"}},
		},
	}

	err := RunWorkflowParallelWithExecutor(wf, exec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	index := make(map[string]int)
	for i, name := range order {
		index[name] = i
	}

	if index["A"] > index["B"] || index["A"] > index["C"] {
		t.Fatal("A should run before B and C")
	}

	if index["B"] > index["D"] || index["C"] > index["D"] {
		t.Fatal("B and C should run before D")
	}
}

// assertError is a simple custom error type for testing.
type assertError struct{}

func (e assertError) Error() string {
	return "test error"
}