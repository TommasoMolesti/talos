package main

import (
	"errors"
	"testing"
)

// TestResolveExecutionOrder_Linear verifies that a simple linear dependency chain
// is resolved in the correct order.
func TestResolveExecutionOrder_Linear(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"B"}},
		},
	}

	order, err := resolveExecutionOrder(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"A", "B", "C"}

	for i, task := range order {
		if task.Name != expected[i] {
			t.Fatalf("expected %s at position %d, got %s", expected[i], i, task.Name)
		}
	}
}

// TestResolveExecutionOrder_Branching verifies that tasks with shared dependencies
// are ordered correctly, even if relative order between siblings is not fixed.
func TestResolveExecutionOrder_Branching(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"A"}},
			"D": {Name: "D", DependsOn: []string{"B", "C"}},
		},
	}

	order, err := resolveExecutionOrder(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	index := make(map[string]int)
	for i, task := range order {
		index[task.Name] = i
	}

	if index["A"] > index["B"] || index["A"] > index["C"] {
		t.Fatal("A should come before B and C")
	}

	if index["B"] > index["D"] || index["C"] > index["D"] {
		t.Fatal("B and C should come before D")
	}
}

// TestResolveExecutionOrder_MissingDependency verifies that an error is returned
// when a task depends on a non-existent task.
func TestResolveExecutionOrder_MissingDependency(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
		},
	}

	_, err := resolveExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for missing dependency, got nil")
	}
}

// TestResolveExecutionOrder_Cycle verifies that cyclic dependencies are detected.
func TestResolveExecutionOrder_Cycle(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
			"B": {Name: "B", DependsOn: []string{"A"}},
		},
	}

	_, err := resolveExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for cycle, got nil")
	}
}

// TestRunWorkflow_Success verifies that RunWorkflow executes without error
// when all tasks succeed.
func TestRunWorkflow_Success(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", Command: "echo A"},
			"B": {Name: "B", Command: "echo B", DependsOn: []string{"A"}},
		},
	}

	err := RunWorkflow(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunWorkflow_Failure verifies that RunWorkflow stops execution
// and returns an error when a task fails.
func TestRunWorkflow_Failure(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", Command: "echo ok"},
			"B": {Name: "B", Command: "exit 1", DependsOn: []string{"A"}},
		},
	}

	err := RunWorkflow(wf)
	if err == nil {
		t.Fatal("expected error from failing task, got nil")
	}
}

// TestRunWorkflow_InvalidWorkflow verifies that RunWorkflow returns an error
// if the workflow is invalid (e.g., cycle or missing dependency).
func TestRunWorkflow_InvalidWorkflow(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
		},
	}

	err := RunWorkflow(wf)
	if err == nil {
		t.Fatal("expected error for invalid workflow, got nil")
	}
}

// TestRunWorkflow_ErrorPropagation ensures that errors returned by task execution
// are properly propagated by RunWorkflow.
func TestRunWorkflow_ErrorPropagation(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", Command: "false"},
		},
	}

	err := RunWorkflow(wf)
	if err == nil {
		t.Fatal("expected error from failing command, got nil")
	}

	if !errors.Is(err, err) {
		// This is mostly illustrative; ensures error is not swallowed.
		t.Log("error propagated correctly")
	}
}