package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func runCLIWithCapturedStderr(t *testing.T, args []string) (int, string) {
	t.Helper()

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stderr = w
	buf := &bytes.Buffer{}

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(buf, r)
		close(done)
	}()

	exitCode := runCLI(args)

	_ = w.Close()
	os.Stderr = orig
	<-done
	_ = r.Close()

	return exitCode, buf.String()
}

func runCLIWithCapturedStdout(t *testing.T, args []string) (int, string) {
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

	exitCode := runCLI(args)

	_ = w.Close()
	os.Stdout = orig
	<-done
	_ = r.Close()

	return exitCode, buf.String()
}

func TestRunCLI_HelpReturnsSuccess(t *testing.T) {
	exitCode, output := runCLIWithCapturedStderr(t, []string{"run", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "-file") || !strings.Contains(output, "-dry-run") || !strings.Contains(output, "-max-concurrency") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestVisualizeCLI_HelpReturnsSuccess(t *testing.T) {
	exitCode, output := runCLIWithCapturedStderr(t, []string{"visualize", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "-file") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestValidateCLI_HelpReturnsSuccess(t *testing.T) {
	exitCode, output := runCLIWithCapturedStderr(t, []string{"validate", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "-file") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestRunCLI_InvalidFlagReturnsFailure(t *testing.T) {
	exitCode, output := runCLIWithCapturedStderr(t, []string{"run", "--bad-flag"})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "flag provided but not defined") {
		t.Fatalf("expected parse error output, got %q", output)
	}
}

func TestRunCmd_UsesCustomWorkflowFile(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "custom.yaml")
	if err := os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	origLoad := loadWorkflowFunc
	origRun := runWorkflowFunc
	defer func() {
		loadWorkflowFunc = origLoad
		runWorkflowFunc = origRun
	}()

	var loadedPath string
	loadWorkflowFunc = func(path string) (*Workflow, error) {
		loadedPath = path
		return origLoad(path)
	}

	var gotWorkflow *Workflow
	var gotOptions RunOptions
	runWorkflowFunc = func(wf *Workflow, opts RunOptions) error {
		gotWorkflow = wf
		gotOptions = opts
		return nil
	}

	if err := runCmd([]string{"--file", workflowPath, "--max-concurrency", "3"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loadedPath != workflowPath {
		t.Fatalf("expected workflow path %q, got %q", workflowPath, loadedPath)
	}

	if gotWorkflow == nil || gotWorkflow.Tasks["demo"] == nil {
		t.Fatalf("expected workflow loaded from custom file")
	}

	if gotOptions.MaxConcurrency != 3 {
		t.Fatalf("expected max concurrency 3, got %d", gotOptions.MaxConcurrency)
	}

	if gotOptions.DryRun {
		t.Fatal("expected dry run to be false")
	}
}

func TestRunCmd_PassesDryRunOption(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "custom.yaml")
	if err := os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	origLoad := loadWorkflowFunc
	origRun := runWorkflowFunc
	defer func() {
		loadWorkflowFunc = origLoad
		runWorkflowFunc = origRun
	}()

	loadWorkflowFunc = func(path string) (*Workflow, error) {
		return origLoad(path)
	}

	var gotOptions RunOptions
	runWorkflowFunc = func(wf *Workflow, opts RunOptions) error {
		gotOptions = opts
		return nil
	}

	if err := runCmd([]string{"--file", workflowPath, "--dry-run"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotOptions.DryRun {
		t.Fatal("expected dry run to be true")
	}
}

func TestValidateCmd_UsesCustomWorkflowFile(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "custom.yaml")
	if err := os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	origLoad := loadWorkflowFunc
	origValidate := validateWorkflowFunc
	defer func() {
		loadWorkflowFunc = origLoad
		validateWorkflowFunc = origValidate
	}()

	var loadedPath string
	loadWorkflowFunc = func(path string) (*Workflow, error) {
		loadedPath = path
		return origLoad(path)
	}

	var gotWorkflow *Workflow
	validateWorkflowFunc = func(wf *Workflow) error {
		gotWorkflow = wf
		return nil
	}

	if err := validateCmd([]string{"--file", workflowPath}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loadedPath != workflowPath {
		t.Fatalf("expected workflow path %q, got %q", workflowPath, loadedPath)
	}

	if gotWorkflow == nil || gotWorkflow.Tasks["demo"] == nil {
		t.Fatalf("expected workflow loaded from custom file")
	}
}

func TestVisualizeCmd_UsesCustomWorkflowFile(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "custom.yaml")
	if err := os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	origLoad := loadWorkflowFunc
	origVisualize := visualizeWorkflowFunc
	defer func() {
		loadWorkflowFunc = origLoad
		visualizeWorkflowFunc = origVisualize
	}()

	var loadedPath string
	loadWorkflowFunc = func(path string) (*Workflow, error) {
		loadedPath = path
		return origLoad(path)
	}

	var gotWorkflow *Workflow
	visualizeWorkflowFunc = func(wf *Workflow) error {
		gotWorkflow = wf
		return nil
	}

	if err := visualizeCmd([]string{"--file", workflowPath}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if loadedPath != workflowPath {
		t.Fatalf("expected workflow path %q, got %q", workflowPath, loadedPath)
	}

	if gotWorkflow == nil || gotWorkflow.Tasks["demo"] == nil {
		t.Fatalf("expected workflow loaded from custom file")
	}
}

func TestRunCLI_VisualizePrintsMermaidGraph(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "visualize.yaml")
	data := "tasks:\n  build:\n    command: \"go build\"\n  test:\n    command: \"go test ./...\"\n    depends_on: [\"build\"]\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	exitCode, output := runCLIWithCapturedStdout(t, []string{"visualize", "--file", workflowPath})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "graph TD") || !strings.Contains(output, "build --> test") {
		t.Fatalf("expected mermaid graph output, got %q", output)
	}
}

func TestValidateCLI_PrintsSuccessMessage(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "validate.yaml")
	data := "tasks:\n  build:\n    command: \"go build\"\n  test:\n    command: \"go test ./...\"\n    depends_on: [\"build\"]\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	exitCode, output := runCLIWithCapturedStdout(t, []string{"validate", "--file", workflowPath})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Workflow is valid (2 tasks)") {
		t.Fatalf("expected validation success output, got %q", output)
	}
}

func TestValidateCLI_InvalidWorkflowReturnsFailure(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "invalid.yaml")
	data := "tasks:\n  build:\n    command: \"go build\"\n    depends_on: [\"missing\"]\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	exitCode, output := runCLIWithCapturedStderr(t, []string{"validate", "--file", workflowPath})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "Validation failed: task build depends on unknown task missing") {
		t.Fatalf("expected validation error output, got %q", output)
	}
}

func TestLoadWorkflow_ParsesTaskTimeout(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "timeout.yaml")
	data := "tasks:\n  slow:\n    command: \"sleep 1\"\n    timeout: 25\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	wf, err := loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Tasks["slow"].TimeoutDuration != 25*time.Second {
		t.Fatalf("expected timeout duration 25s, got %s", wf.Tasks["slow"].TimeoutDuration)
	}
}

func TestLoadWorkflow_RejectsNegativeTaskTimeout(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "invalid-timeout.yaml")
	data := "tasks:\n  slow:\n    command: \"sleep 1\"\n    timeout: -1\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err := loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}

	if !strings.Contains(err.Error(), "timeout must be zero or greater") {
		t.Fatalf("expected invalid timeout error, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNegativeRetries(t *testing.T) {
	tempDir := t.TempDir()
	workflowPath := filepath.Join(tempDir, "invalid-retries.yaml")
	data := "tasks:\n  flaky:\n    command: \"echo nope\"\n    retries: -1\n"
	if err := os.WriteFile(workflowPath, []byte(data), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err := loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected invalid retries error")
	}

	if !strings.Contains(err.Error(), "retries must be zero or greater") {
		t.Fatalf("expected retries validation error, got %v", err)
	}
}
