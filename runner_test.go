package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func captureStdout(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w
	buf := &bytes.Buffer{}

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(buf, r)
		close(done)
	}()

	return buf, func() {
		_ = w.Close()
		os.Stdout = orig
		<-done
		_ = r.Close()
	}
}

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

	runTask = func(_ context.Context, task *Task) error {
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

	runTask = func(_ context.Context, task *Task) error {
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

func TestRunWorkflowParallel_FailureDoesNotDeadlock(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	runTask = func(_ context.Context, task *Task) error {
		if task.Name == "fail" {
			time.Sleep(10 * time.Millisecond)
			return errors.New("boom")
		}

		time.Sleep(30 * time.Millisecond)
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"fail":    {Name: "fail"},
			"succeed": {Name: "succeed"},
		},
	}

	start := time.Now()
	err := RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected workflow error")
	}

	if time.Since(start) > 200*time.Millisecond {
		t.Fatal("workflow took too long, possible deadlock")
	}
}

func TestRunWorkflowParallel_CancelsRunningTasksOnFailure(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	var mu sync.Mutex
	canceled := make(map[string]bool)

	runTask = func(ctx context.Context, task *Task) error {
		if task.Name == "fail" {
			time.Sleep(10 * time.Millisecond)
			return errors.New("boom")
		}

		select {
		case <-ctx.Done():
			mu.Lock()
			canceled[task.Name] = true
			mu.Unlock()
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"fail": {Name: "fail"},
			"slow": {Name: "slow"},
		},
	}

	err := RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected workflow error")
	}

	mu.Lock()
	defer mu.Unlock()
	if !canceled["slow"] {
		t.Fatal("expected slow task to be canceled")
	}
}

func TestRunWorkflowParallel_DryRunDoesNotExecuteTasks(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	runTask = func(_ context.Context, task *Task) error {
		t.Fatalf("did not expect task %s to execute during dry-run", task.Name)
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", Command: "echo A"},
			"B": {Name: "B", Command: "echo B", DependsOn: []string{"A"}},
			"C": {Name: "C", Command: "echo C", DependsOn: []string{"A"}},
			"D": {Name: "D", Command: "echo D", DependsOn: []string{"B", "C"}},
		},
	}

	stdout, restore := captureStdout(t)
	defer restore()

	if err := RunWorkflowParallel(wf, RunOptions{DryRun: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Stage 1: A") {
		t.Fatalf("expected stage 1 output, got %q", output)
	}
	if !strings.Contains(output, "Stage 2: B, C") {
		t.Fatalf("expected stage 2 output, got %q", output)
	}
	if !strings.Contains(output, "Stage 3: D") {
		t.Fatalf("expected stage 3 output, got %q", output)
	}
}

func TestRunWorkflowParallel_TaskTimeoutFailsWorkflow(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	runTask = func(ctx context.Context, task *Task) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"slow": {Name: "slow", Command: "sleep 1", TimeoutDuration: 20 * time.Millisecond},
		},
	}

	err := RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var timeoutErr taskTimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected taskTimeoutError, got %T", err)
	}

	if timeoutErr.task != "slow" {
		t.Fatalf("expected timeout on slow task, got %q", timeoutErr.task)
	}
}

func TestRunWorkflowParallel_RetriesTaskUntilSuccess(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	attempts := 0
	runTask = func(_ context.Context, task *Task) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient failure")
		}
		return nil
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"flaky": {Name: "flaky", Command: "echo flaky", Retries: 2},
		},
	}

	if err := RunWorkflowParallel(wf, RunOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRunWorkflowParallel_RetriesExhausted(t *testing.T) {
	orig := runTask
	defer func() { runTask = orig }()

	attempts := 0
	runTask = func(_ context.Context, task *Task) error {
		attempts++
		return errors.New("still failing")
	}

	wf := &Workflow{
		Tasks: map[string]*Task{
			"flaky": {Name: "flaky", Command: "echo flaky", Retries: 2},
		},
	}

	err := RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected workflow error")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
