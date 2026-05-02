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

// captureStdout captures stdout until the returned restore function is called.
func captureStdout(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	var orig *os.File = os.Stdout
	var r *os.File
	var w *os.File
	var err error
	r, w, err = os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w
	var buf *bytes.Buffer = &bytes.Buffer{}

	var done chan struct{} = make(chan struct{})
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
	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"B"}},
		},
	}

	var err error = validateExecutionOrder(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyExecutionOrder_MissingDependency(t *testing.T) {
	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
		},
	}

	var err error = validateExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for missing dependency")
	}
}

func TestVerifyExecutionOrder_Cycle(t *testing.T) {
	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", DependsOn: []string{"B"}},
			"B": {Name: "B", DependsOn: []string{"A"}},
		},
	}

	var err error = validateExecutionOrder(wf)
	if err == nil {
		t.Fatal("expected error for cycle")
	}
}

// --- EXECUTION TESTS ---

func TestRunWorkflowParallel_AllTasksExecuted(t *testing.T) {
	var mu sync.Mutex
	var executed map[string]bool = make(map[string]bool)

	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	runTask = func(_ context.Context, task *Task) error {
		mu.Lock()
		executed[task.Name] = true
		mu.Unlock()
		return nil
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B", DependsOn: []string{"A"}},
			"C": {Name: "C", DependsOn: []string{"A"}},
			"D": {Name: "D", DependsOn: []string{"B", "C"}},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{})
	if err != nil {
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
	var running int = 0
	var maxSeen int = 0

	var orig func(context.Context, *Task) error = runTask
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

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A"},
			"B": {Name: "B"},
			"C": {Name: "C"},
			"D": {Name: "D"},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{MaxConcurrency: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxSeen > 2 {
		t.Fatalf("expected max concurrency 2, got %d", maxSeen)
	}
}

func TestRunWorkflowParallel_FailureDoesNotDeadlock(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	runTask = func(_ context.Context, task *Task) error {
		if task.Name == "fail" {
			time.Sleep(10 * time.Millisecond)
			return errors.New("boom")
		}

		time.Sleep(30 * time.Millisecond)
		return nil
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"fail":    {Name: "fail"},
			"succeed": {Name: "succeed"},
		},
	}

	var start time.Time = time.Now()
	var err error = RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected workflow error")
	}

	if time.Since(start) > 200*time.Millisecond {
		t.Fatal("workflow took too long, possible deadlock")
	}
}

func TestRunWorkflowParallel_CancelsRunningTasksOnFailure(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	var mu sync.Mutex
	var canceled map[string]bool = make(map[string]bool)

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

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"fail": {Name: "fail"},
			"slow": {Name: "slow"},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{})
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
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	runTask = func(_ context.Context, task *Task) error {
		t.Fatalf("did not expect task %s to execute during dry-run", task.Name)
		return nil
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"A": {Name: "A", Description: "Prepare workspace", Command: "echo A"},
			"B": {Name: "B", Command: "echo B", DependsOn: []string{"A"}},
			"C": {Name: "C", Command: "echo C", DependsOn: []string{"A"}},
			"D": {Name: "D", Command: "echo D", DependsOn: []string{"B", "C"}},
		},
	}

	var stdout *bytes.Buffer
	var restore func()
	stdout, restore = captureStdout(t)
	var err error = RunWorkflowParallel(wf, RunOptions{DryRun: true})
	if err != nil {
		restore()
		t.Fatalf("unexpected error: %v", err)
	}
	restore()

	var output string = stdout.String()
	if !strings.Contains(output, "Stage 1: A") {
		t.Fatalf("expected stage 1 output, got %q", output)
	}
	if !strings.Contains(output, "A - Prepare workspace: echo A") {
		t.Fatalf("expected description in dry-run output, got %q", output)
	}
	if !strings.Contains(output, "Stage 2: B, C") {
		t.Fatalf("expected stage 2 output, got %q", output)
	}
	if !strings.Contains(output, "Stage 3: D") {
		t.Fatalf("expected stage 3 output, got %q", output)
	}
}

func TestRunWorkflowParallel_TaskTimeoutFailsWorkflow(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	runTask = func(ctx context.Context, task *Task) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"slow": {Name: "slow", Command: "sleep 1", TimeoutDuration: 20 * time.Millisecond},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{})
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
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	var attempts int = 0
	runTask = func(_ context.Context, task *Task) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient failure")
		}
		return nil
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"flaky": {Name: "flaky", Description: "Retry transient command", Command: "echo flaky", Retries: 2},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRunWorkflowParallel_PrintsSummaryWithRetries(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	var attempts int = 0
	runTask = func(_ context.Context, task *Task) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient failure")
		}
		return nil
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"flaky": {Name: "flaky", Description: "Retry transient command", Command: "echo flaky", Retries: 2},
		},
	}

	var stdout *bytes.Buffer
	var restore func()
	stdout, restore = captureStdout(t)
	var err error = RunWorkflowParallel(wf, RunOptions{})
	if err != nil {
		restore()
		t.Fatalf("unexpected error: %v", err)
	}
	restore()

	var output string = stdout.String()
	if !strings.Contains(output, "[talos] Summary") {
		t.Fatalf("expected summary output, got %q", output)
	}
	if !strings.Contains(output, "total=1 success=1 failed=0 timed_out=0 canceled=0 skipped=0") {
		t.Fatalf("expected summary counts, got %q", output)
	}
	if !strings.Contains(output, "retries: flaky - Retry transient command (2 retries)") {
		t.Fatalf("expected retry details, got %q", output)
	}
}

func TestRunWorkflowParallel_RetriesExhausted(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	var attempts int = 0
	runTask = func(_ context.Context, task *Task) error {
		attempts++
		return errors.New("still failing")
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"flaky": {Name: "flaky", Command: "echo flaky", Retries: 2},
		},
	}

	var err error = RunWorkflowParallel(wf, RunOptions{})
	if err == nil {
		t.Fatal("expected workflow error")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestRunWorkflowParallel_PrintsSummaryWithTimeoutsAndSkipsOnFailure(t *testing.T) {
	var orig func(context.Context, *Task) error = runTask
	defer func() { runTask = orig }()

	runTask = func(ctx context.Context, task *Task) error {
		switch task.Name {
		case "slow":
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		case "blocked":
			t.Fatal("did not expect blocked task to run")
			return nil
		default:
			return nil
		}
	}

	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"slow":    {Name: "slow", Description: "Wait too long", Command: "sleep 1", TimeoutDuration: 20 * time.Millisecond},
			"blocked": {Name: "blocked", Description: "Runs after slow", Command: "echo blocked", DependsOn: []string{"slow"}},
		},
	}

	var stdout *bytes.Buffer
	var restore func()
	stdout, restore = captureStdout(t)
	var err error = RunWorkflowParallel(wf, RunOptions{})
	restore()
	if err == nil {
		t.Fatal("expected workflow error")
	}

	var output string = stdout.String()
	if !strings.Contains(output, "[talos] Summary") {
		t.Fatalf("expected summary output, got %q", output)
	}
	if !strings.Contains(output, "total=2 success=0 failed=0 timed_out=1 canceled=0 skipped=1") {
		t.Fatalf("expected timeout and skip counts, got %q", output)
	}
	if !strings.Contains(output, "timeouts: slow - Wait too long (20ms)") {
		t.Fatalf("expected timeout details, got %q", output)
	}
	if !strings.Contains(output, "skipped: blocked - Runs after slow") {
		t.Fatalf("expected skipped task details, got %q", output)
	}
	if !strings.Contains(output, "[talos] Failed in") {
		t.Fatalf("expected failed final line, got %q", output)
	}
}

func TestRunTask_UsesTaskWorkingDirAndEnv(t *testing.T) {
	var tempDir string = t.TempDir()
	var task *Task = &Task{
		Name:    "demo",
		Command: "pwd && printf '%s\\n' \"$TASK_MODE\"",
		Cwd:     tempDir,
		Env: map[string]string{
			"TASK_MODE": "enabled",
		},
	}

	var stdout *bytes.Buffer
	var restore func()
	stdout, restore = captureStdout(t)
	var err error = runTask(context.Background(), task)
	restore()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output string = stdout.String()
	if !strings.Contains(output, tempDir) {
		t.Fatalf("expected task output to include working dir %q, got %q", tempDir, output)
	}
	if !strings.Contains(output, "enabled") {
		t.Fatalf("expected task output to include env var, got %q", output)
	}
}
