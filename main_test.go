package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestRunCLI_HelpReturnsSuccess(t *testing.T) {
	exitCode, output := runCLIWithCapturedStderr(t, []string{"run", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "-file") || !strings.Contains(output, "-max-concurrency") {
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
}
