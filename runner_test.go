package main

import (
	"sync"
	"testing"
	"time"
)

// --- VALIDATION TESTS ---

func TestVerifyExecutionOrder_Linear(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"B"}},
		},
	}

	if err := validateExecutionOrder(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyExecutionOrder_MissingDependency(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
		},
	}

	if err := validateExecutionOrder(wf); err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestVerifyExecutionOrder_Cycle(t *testing.T) {
	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
			"B": {Name: "B", DependsOn: []string{"A"}},
		},
	}

	if err := validateExecutionOrder(wf); err == nil {
		t.Fatal("expected error for cycle")
	}
}

// --- EXECUTION TESTS ---

func TestRunWorkflowParallel_AllTasksExecuted(t *testing.T) {
	var mu sync.Mutex
	executed := make(map[string]bool)

	orig := runTask
	defer func() { runTask = orig }()

	runTask = func(task *Task) error {
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

	if err := RunWorkflowParallel(wf, RunOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for name := range wf.Tasks {
		if !executed[name] {
			t.Fatalf("task %s not executed", name)
		}
	}
}

func TestRunWorkflowParallel_MaxConcurrency(t *testing.T) {
	var mu sync.Mutex
	running := 0
	maxSeen := 0

	orig := runTask
	defer func() { runTask = orig }()

	runTask = func(task *Task) error {
		mu.Lock()
		running++
		if running > maxSeen {
			maxSeen = running
		}
		mu.Unlock()

		time.Sleep(50 * time.Millisecond)

		mu.Lock()
		running--
		mu.Unlock()

		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B"},
			"C": {Name: "C"},
			"D": {Name: "D"},
		},
	}

	if err := RunWorkflowParallel(wf, RunOptions{MaxConcurrency: 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxSeen > 2 {
		t.Fatalf("expected max concurrency 2, got %d", maxSeen)
	}
}