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

	var orig *os.File = os.Stderr
	var r *os.File
	var w *os.File
	var err error
	r, w, err = os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stderr = w
	var buf *bytes.Buffer = &bytes.Buffer{}

	var done chan struct{} = make(chan struct{})
	go func() {
		_, _ = io.Copy(buf, r)
		close(done)
	}()

	var exitCode int = runCLI(args)

	_ = w.Close()
	os.Stderr = orig
	<-done
	_ = r.Close()

	return exitCode, buf.String()
}

func runCLIWithCapturedStdout(t *testing.T, args []string) (int, string) {
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

	var exitCode int = runCLI(args)

	_ = w.Close()
	os.Stdout = orig
	<-done
	_ = r.Close()

	return exitCode, buf.String()
}

func TestRunCLI_HelpReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"run", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage: talos run [flags]") || !strings.Contains(output, "talos run --dry-run --target test") {
		t.Fatalf("expected usage examples, got %q", output)
	}

	if !strings.Contains(output, "-file") || !strings.Contains(output, "-dry-run") || !strings.Contains(output, "-max-concurrency") || !strings.Contains(output, "-target") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestVisualizeCLI_HelpReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"visualize", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage: talos visualize [flags]") || !strings.Contains(output, "talos visualize --file ./workflows/dev.yaml") {
		t.Fatalf("expected help output, got %q", output)
	}

	if !strings.Contains(output, "-file") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestValidateCLI_HelpReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"validate", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage: talos validate [flags]") || !strings.Contains(output, "talos validate --file ./workflows/dev.yaml") {
		t.Fatalf("expected help output, got %q", output)
	}

	if !strings.Contains(output, "-file") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestInitCLI_HelpReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"init", "-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage: talos init [flags]") || !strings.Contains(output, "talos init --file ./workflows/dev.yaml") {
		t.Fatalf("expected help output, got %q", output)
	}

	if !strings.Contains(output, "-file") || !strings.Contains(output, "-force") {
		t.Fatalf("expected help output, got %q", output)
	}
}

func TestCLI_RootHelpFlagReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"-h"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Usage:\n  talos <command> [flags]") {
		t.Fatalf("expected root usage output, got %q", output)
	}
	if !strings.Contains(output, "Commands:") || !strings.Contains(output, "run        Execute a workflow") {
		t.Fatalf("expected command guidance, got %q", output)
	}
}

func TestCLI_HelpCommandReturnsSuccess(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"help"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Use \"talos <command> -h\" for command-specific help.") {
		t.Fatalf("expected root help hint, got %q", output)
	}
}

func TestCLI_VersionCommandPrintsBuildMetadata(t *testing.T) {
	var origVersion string = version
	var origCommit string = commit
	var origDate string = date
	defer func() {
		version = origVersion
		commit = origCommit
		date = origDate
	}()

	version = "1.2.3"
	commit = "abc123"
	date = "2026-05-02T10:11:12Z"

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"version"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "talos 1.2.3") || !strings.Contains(output, "commit: abc123") || !strings.Contains(output, "built: 2026-05-02T10:11:12Z") {
		t.Fatalf("expected version output, got %q", output)
	}
}

func TestCLI_NoArgsPrintsRootUsage(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, nil)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "Talos executes local task workflows") {
		t.Fatalf("expected root usage, got %q", output)
	}
}

func TestRunCLI_InvalidFlagReturnsFailure(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"run", "--bad-flag"})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "flag provided but not defined") {
		t.Fatalf("expected parse error output, got %q", output)
	}
}

func TestCLI_UnknownCommandPrintsGuidance(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"unknown"})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "Unknown command: unknown") || !strings.Contains(output, "Commands:") {
		t.Fatalf("expected unknown command guidance, got %q", output)
	}
}

func TestInitCmd_WritesStarterWorkflow(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "workflows", "dev.yaml")

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"init", "--file", workflowPath})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Created starter workflow") {
		t.Fatalf("expected success output, got %q", output)
	}

	var data []byte
	var err error
	data, err = os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read starter workflow: %v", err)
	}

	if string(data) != starterWorkflow {
		t.Fatalf("expected starter workflow, got %q", string(data))
	}
}

func TestInitCmd_RefusesToOverwriteExistingWorkflow(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "talos.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks: {}\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"init", "--file", workflowPath})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "already exists; use --force to overwrite") {
		t.Fatalf("expected overwrite guidance, got %q", output)
	}
}

func TestInitCmd_ForceOverwritesExistingWorkflow(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "talos.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks: {}\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	err = initCmd([]string{"--file", workflowPath, "--force"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data []byte
	data, err = os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read starter workflow: %v", err)
	}

	if string(data) != starterWorkflow {
		t.Fatalf("expected starter workflow, got %q", string(data))
	}
}

func TestRunCmd_UsesCustomWorkflowFile(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "custom.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var origLoad func(string) (*Workflow, error) = loadWorkflowFunc
	var origRun func(*Workflow, RunOptions) error = runWorkflowFunc
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

	err = runCmd([]string{"--file", workflowPath, "--max-concurrency", "3"})
	if err != nil {
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
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "custom.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var origLoad func(string) (*Workflow, error) = loadWorkflowFunc
	var origRun func(*Workflow, RunOptions) error = runWorkflowFunc
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

	err = runCmd([]string{"--file", workflowPath, "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotOptions.DryRun {
		t.Fatal("expected dry run to be true")
	}
}

func TestRunCmd_TargetFiltersWorkflowToDependencies(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "target.yaml")
	var data string = strings.Join([]string{
		"tasks:",
		"  install:",
		"    command: \"npm install\"",
		"  lint:",
		"    command: \"npm run lint\"",
		"  build:",
		"    command: \"npm run build\"",
		"    depends_on: [\"install\"]",
		"  test:",
		"    command: \"npm test\"",
		"    depends_on: [\"build\"]",
	}, "\n")
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var origRun func(*Workflow, RunOptions) error = runWorkflowFunc
	defer func() { runWorkflowFunc = origRun }()

	var gotWorkflow *Workflow
	runWorkflowFunc = func(wf *Workflow, opts RunOptions) error {
		gotWorkflow = wf
		return nil
	}

	err = runCmd([]string{"--file", workflowPath, "--target", "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotWorkflow == nil {
		t.Fatal("expected workflow to be passed to runner")
	}

	if len(gotWorkflow.Tasks) != 3 {
		t.Fatalf("expected 3 tasks in targeted workflow, got %d", len(gotWorkflow.Tasks))
	}

	for _, name := range []string{"install", "build", "test"} {
		if gotWorkflow.Tasks[name] == nil {
			t.Fatalf("expected task %s in targeted workflow", name)
		}
	}

	if gotWorkflow.Tasks["lint"] != nil {
		t.Fatal("did not expect unrelated task lint in targeted workflow")
	}
}

func TestRunCmd_TargetMissingTaskReturnsError(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "target.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	err = runCmd([]string{"--file", workflowPath, "--target", "missing"})
	if err == nil {
		t.Fatal("expected missing target error")
	}

	if !strings.Contains(err.Error(), "target task missing not found") {
		t.Fatalf("expected missing target error, got %v", err)
	}
}

func TestValidateCmd_UsesCustomWorkflowFile(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "custom.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var origLoad func(string) (*Workflow, error) = loadWorkflowFunc
	var origValidate func(*Workflow) error = validateWorkflowFunc
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

	err = validateCmd([]string{"--file", workflowPath})
	if err != nil {
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
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "custom.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks:\n  demo:\n    command: \"echo demo\"\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var origLoad func(string) (*Workflow, error) = loadWorkflowFunc
	var origVisualize func(*Workflow) error = visualizeWorkflowFunc
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

	err = visualizeCmd([]string{"--file", workflowPath})
	if err != nil {
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
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "visualize.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build\"\n  test:\n    command: \"go test ./...\"\n    depends_on: [\"build\"]\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"visualize", "--file", workflowPath})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "graph TD") || !strings.Contains(output, "build --> test") {
		t.Fatalf("expected mermaid graph output, got %q", output)
	}
}

func TestRunCLI_TargetDryRunPrintsOnlyRequiredTasks(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "target-dry-run.yaml")
	var data string = strings.Join([]string{
		"tasks:",
		"  install:",
		"    command: \"npm install\"",
		"  lint:",
		"    command: \"npm run lint\"",
		"  build:",
		"    command: \"npm run build\"",
		"    depends_on: [\"install\"]",
		"  test:",
		"    command: \"npm test\"",
		"    depends_on: [\"build\"]",
	}, "\n")
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"run", "--file", workflowPath, "--target", "test", "--dry-run"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if strings.Contains(output, "lint") {
		t.Fatalf("did not expect unrelated task in dry-run output, got %q", output)
	}
	if !strings.Contains(output, "Stage 1: install") || !strings.Contains(output, "Stage 2: build") || !strings.Contains(output, "Stage 3: test") {
		t.Fatalf("expected targeted dry-run output, got %q", output)
	}
}

func TestValidateCLI_PrintsSuccessMessage(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "validate.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build\"\n  test:\n    command: \"go test ./...\"\n    depends_on: [\"build\"]\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"validate", "--file", workflowPath})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(output, "Workflow is valid (2 tasks)") {
		t.Fatalf("expected validation success output, got %q", output)
	}
}

func TestValidateCLI_InvalidWorkflowReturnsFailure(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "invalid.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build\"\n    depends_on: [\"missing\"]\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"validate", "--file", workflowPath})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "Validation failed: task build depends on unknown task missing") {
		t.Fatalf("expected validation error output, got %q", output)
	}
}

func TestLoadWorkflow_ParsesTaskTimeout(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "timeout.yaml")
	var data string = "tasks:\n  slow:\n    command: \"sleep 1\"\n    timeout: 25\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var wf *Workflow
	wf, err = loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Tasks["slow"].TimeoutDuration != 25*time.Second {
		t.Fatalf("expected timeout duration 25s, got %s", wf.Tasks["slow"].TimeoutDuration)
	}
}

func TestLoadWorkflow_ResolvesTaskWorkingDirAndEnv(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowDir string = filepath.Join(tempDir, "config")
	var projectDir string = filepath.Join(tempDir, "project")
	var err error = os.MkdirAll(workflowDir, 0o755)
	if err != nil {
		t.Fatalf("create workflow dir: %v", err)
	}
	err = os.MkdirAll(projectDir, 0o755)
	if err != nil {
		t.Fatalf("create project dir: %v", err)
	}

	var workflowPath string = filepath.Join(workflowDir, "talos.yaml")
	var data string = strings.Join([]string{
		"tasks:",
		"  demo:",
		"    command: \"pwd\"",
		"    cwd: \"../project\"",
		"    env:",
		"      APP_MODE: \"dev\"",
	}, "\n")
	err = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var wf *Workflow
	wf, err = loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var task *Task = wf.Tasks["demo"]
	if task.WorkingDir != projectDir {
		t.Fatalf("expected working dir %q, got %q", projectDir, task.WorkingDir)
	}

	if task.Env["APP_MODE"] != "dev" {
		t.Fatalf("expected APP_MODE env to be parsed, got %#v", task.Env)
	}
}

func TestLoadWorkflow_RejectsNegativeTaskTimeout(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "invalid-timeout.yaml")
	var data string = "tasks:\n  slow:\n    command: \"sleep 1\"\n    timeout: -1\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}

	if !strings.Contains(err.Error(), "timeout must be zero or greater") {
		t.Fatalf("expected invalid timeout error, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNegativeRetries(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "invalid-retries.yaml")
	var data string = "tasks:\n  flaky:\n    command: \"echo nope\"\n    retries: -1\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected invalid retries error")
	}

	if !strings.Contains(err.Error(), "retries must be zero or greater") {
		t.Fatalf("expected retries validation error, got %v", err)
	}
}
