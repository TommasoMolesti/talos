package main

import (
	"bytes"
	"io"
	"os"
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

	if !strings.Contains(output, "-max-concurrency") {
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
