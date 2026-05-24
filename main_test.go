package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// runCLIWithCapturedStderr runs the CLI and returns anything written to stderr.
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

// runCLIWithCapturedStdout runs the CLI and returns anything written to stdout.
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

	if !strings.Contains(output, "-file") || !strings.Contains(output, "-dry-run") || !strings.Contains(output, "-max-concurrency") || !strings.Contains(output, "-quiet") || !strings.Contains(output, "-summary") || !strings.Contains(output, "-target") || !strings.Contains(output, "-verbose") {
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

func TestRunCLI_RejectsUnexpectedCommandArguments(t *testing.T) {
	var cases []struct {
		name string
		args []string
		want string
	} = []struct {
		name string
		args []string
		want string
	}{
		{name: "init", args: []string{"init", "extra"}, want: "Init failed: unexpected argument \"extra\""},
		{name: "run", args: []string{"run", "extra"}, want: "Execution failed: unexpected argument \"extra\""},
		{name: "validate", args: []string{"validate", "extra"}, want: "Validation failed: unexpected argument \"extra\""},
		{name: "visualize", args: []string{"visualize", "extra"}, want: "Visualization failed: unexpected argument \"extra\""},
		{name: "version", args: []string{"version", "extra"}, want: "Version failed: unexpected argument \"extra\""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var exitCode int
			var output string
			exitCode, output = runCLIWithCapturedStderr(t, tc.args)
			if exitCode != 1 {
				t.Fatalf("expected exit code 1, got %d", exitCode)
			}
			if !strings.Contains(output, tc.want) {
				t.Fatalf("expected unexpected argument output %q, got %q", tc.want, output)
			}
		})
	}
}

func TestRunCLI_RejectsMultipleUnexpectedCommandArguments(t *testing.T) {
	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"run", "first", "second"})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(output, "Execution failed: unexpected arguments: first, second") {
		t.Fatalf("expected multiple unexpected arguments output, got %q", output)
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

func TestRunCmd_PassesQuietOption(t *testing.T) {
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

	err = runCmd([]string{"--file", workflowPath, "--quiet"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotOptions.Quiet {
		t.Fatal("expected quiet to be true")
	}
}

func TestRunCmd_PassesVerboseOption(t *testing.T) {
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

	err = runCmd([]string{"--file", workflowPath, "--verbose"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gotOptions.Verbose {
		t.Fatal("expected verbose to be true")
	}
}

func TestRunCmd_PassesJSONSummaryOption(t *testing.T) {
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

	err = runCmd([]string{"--file", workflowPath, "--summary", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotOptions.SummaryFormat != "json" {
		t.Fatalf("expected JSON summary format, got %q", gotOptions.SummaryFormat)
	}
}

func TestRunCmd_RejectsQuietAndVerboseTogether(t *testing.T) {
	var err error = runCmd([]string{"--quiet", "--verbose"})
	if err == nil {
		t.Fatal("expected quiet and verbose conflict")
	}
	if !strings.Contains(err.Error(), "--quiet and --verbose cannot be used together") {
		t.Fatalf("expected quiet and verbose conflict, got %v", err)
	}
}

func TestRunCmd_RejectsNegativeMaxConcurrency(t *testing.T) {
	var err error = runCmd([]string{"--max-concurrency", "-1"})
	if err == nil {
		t.Fatal("expected negative max concurrency error")
	}
	if !strings.Contains(err.Error(), "--max-concurrency must be zero or greater") {
		t.Fatalf("expected max concurrency validation error, got %v", err)
	}
}

func TestRunCmd_RejectsUnknownSummaryFormat(t *testing.T) {
	var err error = runCmd([]string{"--summary", "xml"})
	if err == nil {
		t.Fatal("expected unknown summary format error")
	}
	if !strings.Contains(err.Error(), "--summary must be human or json") {
		t.Fatalf("expected summary format error, got %v", err)
	}
}

func TestRunCmd_RejectsJSONSummaryAndVerboseTogether(t *testing.T) {
	var err error = runCmd([]string{"--summary", "json", "--verbose"})
	if err == nil {
		t.Fatal("expected JSON summary and verbose conflict")
	}
	if !strings.Contains(err.Error(), "--summary json and --verbose cannot be used together") {
		t.Fatalf("expected JSON summary and verbose conflict, got %v", err)
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

func TestRunCLI_JSONSummaryPrintsParseableSummary(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "json-summary.yaml")
	var data string = "tasks:\n  demo:\n    description: \"Demo task\"\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStdout(t, []string{"run", "--file", workflowPath, "--summary", "json"})
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; output=%q", exitCode, output)
	}
	if strings.Contains(output, "[talos]") || strings.Contains(output, "[demo] demo") {
		t.Fatalf("expected JSON-only output, got %q", output)
	}

	var summary struct {
		Success bool `json:"success"`
		Counts  map[string]int
		Tasks   []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Status      string `json:"status"`
		} `json:"tasks"`
	}
	err = json.Unmarshal([]byte(output), &summary)
	if err != nil {
		t.Fatalf("parse JSON summary: %v; output=%q", err, output)
	}
	if !summary.Success || summary.Counts["success"] != 1 || len(summary.Tasks) != 1 || summary.Tasks[0].Name != "demo" || summary.Tasks[0].Description != "Demo task" || summary.Tasks[0].Status != "success" {
		t.Fatalf("expected successful JSON summary, got %#v", summary)
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

	if !strings.Contains(output, "Validation failed:") || !strings.Contains(output, "task build depends on unknown task missing") {
		t.Fatalf("expected validation error output, got %q", output)
	}
	if !strings.Contains(output, workflowPath+":4:18") {
		t.Fatalf("expected validation error location, got %q", output)
	}
}

func TestValidateCLI_MissingTaskCommandReturnsFailure(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "missing-command.yaml")
	var data string = "tasks:\n  build:\n    description: \"Compile the app\"\n"
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

	if !strings.Contains(output, "task build command is required") {
		t.Fatalf("expected missing command validation error, got %q", output)
	}
	if !strings.Contains(output, workflowPath+":2:3") {
		t.Fatalf("expected missing command error location, got %q", output)
	}
}

func TestValidateCLI_DuplicateDependencyReturnsFailure(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "duplicate-dependency.yaml")
	var data string = strings.Join([]string{
		"tasks:",
		"  test:",
		"    command: \"go test ./...\"",
		"  build:",
		"    command: \"go build ./...\"",
		"    depends_on: [\"test\", \"test\"]",
	}, "\n")
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

	if !strings.Contains(output, "task build depends on test more than once") {
		t.Fatalf("expected duplicate dependency validation error, got %q", output)
	}
	if !strings.Contains(output, workflowPath+":6:26") {
		t.Fatalf("expected duplicate dependency location, got %q", output)
	}
}

func TestValidateCLI_EmptyWorkflowReturnsFailure(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "empty.yaml")
	var err error = os.WriteFile(workflowPath, []byte("tasks: {}\n"), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var exitCode int
	var output string
	exitCode, output = runCLIWithCapturedStderr(t, []string{"validate", "--file", workflowPath})
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(output, "workflow must define at least one task") {
		t.Fatalf("expected empty workflow validation error, got %q", output)
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

func TestLoadWorkflow_ParsesTaskDescription(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "description.yaml")
	var data string = "tasks:\n  build:\n    description: \"Compile the app\"\n    command: \"go build\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var wf *Workflow
	wf, err = loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Tasks["build"].Description != "Compile the app" {
		t.Fatalf("expected description to be parsed, got %q", wf.Tasks["build"].Description)
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

func TestLoadWorkflow_AppliesWorkflowDefaults(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowDir string = filepath.Join(tempDir, "config")
	var appDir string = filepath.Join(tempDir, "app")
	var overrideDir string = filepath.Join(tempDir, "override")
	var err error = os.MkdirAll(workflowDir, 0o755)
	if err != nil {
		t.Fatalf("create workflow dir: %v", err)
	}
	err = os.MkdirAll(appDir, 0o755)
	if err != nil {
		t.Fatalf("create app dir: %v", err)
	}
	err = os.MkdirAll(overrideDir, 0o755)
	if err != nil {
		t.Fatalf("create override dir: %v", err)
	}

	var workflowPath string = filepath.Join(workflowDir, "talos.yaml")
	var data string = strings.Join([]string{
		"defaults:",
		"  cwd: \"../app\"",
		"  shell: \"bash\"",
		"  env:",
		"    APP_ENV: \"dev\"",
		"    SHARED: \"default\"",
		"  retries: 2",
		"  timeout: 30",
		"tasks:",
		"  inherited:",
		"    command: \"go test ./...\"",
		"  override:",
		"    command: \"go build\"",
		"    shell: \"zsh\"",
		"    cwd: \"../override\"",
		"    env:",
		"      SHARED: \"task\"",
		"      TASK_ONLY: \"yes\"",
		"    retries: 0",
		"    timeout: 0",
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

	var inherited *Task = wf.Tasks["inherited"]
	if inherited.WorkingDir != appDir {
		t.Fatalf("expected inherited working dir %q, got %q", appDir, inherited.WorkingDir)
	}
	if inherited.Shell != "bash" {
		t.Fatalf("expected inherited shell bash, got %q", inherited.Shell)
	}
	if inherited.Env["APP_ENV"] != "dev" || inherited.Env["SHARED"] != "default" {
		t.Fatalf("expected inherited env defaults, got %#v", inherited.Env)
	}
	if inherited.Retries != 2 {
		t.Fatalf("expected inherited retries 2, got %d", inherited.Retries)
	}
	if inherited.TimeoutDuration != 30*time.Second {
		t.Fatalf("expected inherited timeout 30s, got %s", inherited.TimeoutDuration)
	}

	var override *Task = wf.Tasks["override"]
	if override.WorkingDir != overrideDir {
		t.Fatalf("expected override working dir %q, got %q", overrideDir, override.WorkingDir)
	}
	if override.Shell != "zsh" {
		t.Fatalf("expected task shell override zsh, got %q", override.Shell)
	}
	if override.Env["APP_ENV"] != "dev" || override.Env["SHARED"] != "task" || override.Env["TASK_ONLY"] != "yes" {
		t.Fatalf("expected merged env with task overrides, got %#v", override.Env)
	}
	if override.Retries != 0 {
		t.Fatalf("expected task retries override 0, got %d", override.Retries)
	}
	if override.TimeoutDuration != 0 {
		t.Fatalf("expected task timeout override 0, got %s", override.TimeoutDuration)
	}
}

func TestLoadWorkflow_ParsesShellExample(t *testing.T) {
	var workflowPath string = filepath.Join("examples", "shell.yaml")

	var wf *Workflow
	var err error
	wf, err = loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var defaultShell *Task = wf.Tasks["default-shell"]
	if defaultShell == nil {
		t.Fatal("expected default-shell task")
	}
	if defaultShell.Shell != "bash" {
		t.Fatalf("expected default-shell to inherit bash, got %q", defaultShell.Shell)
	}

	var taskShell *Task = wf.Tasks["task-shell"]
	if taskShell == nil {
		t.Fatal("expected task-shell task")
	}
	if taskShell.Shell != "zsh" {
		t.Fatalf("expected task-shell to override zsh, got %q", taskShell.Shell)
	}
}

func TestLoadWorkflow_TrimsShellValuesAndIgnoresBlankOverride(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "shell.yaml")
	var data string = strings.Join([]string{
		"defaults:",
		"  shell: \"  bash  \"",
		"tasks:",
		"  inherited:",
		"    command: \"echo inherited\"",
		"  blank:",
		"    command: \"echo blank\"",
		"    shell: \"   \"",
		"  override:",
		"    command: \"echo override\"",
		"    shell: \"  zsh  \"",
	}, "\n")
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	var wf *Workflow
	wf, err = loadWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wf.Tasks["inherited"].Shell != "bash" {
		t.Fatalf("expected inherited shell bash, got %q", wf.Tasks["inherited"].Shell)
	}
	if wf.Tasks["blank"].Shell != "bash" {
		t.Fatalf("expected blank task shell to inherit bash, got %q", wf.Tasks["blank"].Shell)
	}
	if wf.Tasks["override"].Shell != "zsh" {
		t.Fatalf("expected override shell zsh, got %q", wf.Tasks["override"].Shell)
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
	if !strings.Contains(err.Error(), workflowPath+":4:5") {
		t.Fatalf("expected invalid timeout location, got %v", err)
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

func TestLoadWorkflow_RejectsNegativeDefaults(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "invalid-defaults.yaml")
	var data string = "defaults:\n  retries: -1\ntasks:\n  demo:\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected invalid defaults error")
	}

	if !strings.Contains(err.Error(), "defaults retries must be zero or greater") {
		t.Fatalf("expected defaults validation error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":2:3") {
		t.Fatalf("expected defaults validation location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsUnknownTopLevelField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "unknown-top-level.yaml")
	var data string = "version: 1\ntasks:\n  demo:\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected unknown top-level field error")
	}

	if !strings.Contains(err.Error(), "unsupported top-level field \"version\"") {
		t.Fatalf("expected unknown top-level field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":1:1") {
		t.Fatalf("expected unknown top-level field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsDuplicateTopLevelField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "duplicate-top-level.yaml")
	var data string = "tasks:\n  first:\n    command: \"echo first\"\ntasks:\n  second:\n    command: \"echo second\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected duplicate top-level field error")
	}

	if !strings.Contains(err.Error(), "duplicate top-level field \"tasks\"") {
		t.Fatalf("expected duplicate top-level field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:1") {
		t.Fatalf("expected duplicate top-level field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsUnknownDefaultField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "unknown-default.yaml")
	var data string = "defaults:\n  path: \".\"\ntasks:\n  demo:\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected unknown default field error")
	}

	if !strings.Contains(err.Error(), "unsupported field \"path\" in defaults") {
		t.Fatalf("expected unknown default field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":2:3") {
		t.Fatalf("expected unknown default field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNonMappingDefaults(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "non-mapping-defaults.yaml")
	var data string = "defaults: nope\ntasks:\n  demo:\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected non-mapping defaults error")
	}

	if !strings.Contains(err.Error(), "defaults must be a mapping") {
		t.Fatalf("expected non-mapping defaults error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":1:11") {
		t.Fatalf("expected non-mapping defaults location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsDuplicateDefaultField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "duplicate-default.yaml")
	var data string = "defaults:\n  timeout: 30\n  timeout: 60\ntasks:\n  demo:\n    command: \"echo demo\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected duplicate default field error")
	}

	if !strings.Contains(err.Error(), "duplicate field \"timeout\" in defaults") {
		t.Fatalf("expected duplicate default field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":3:3") {
		t.Fatalf("expected duplicate default field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNonMappingTasks(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "non-mapping-tasks.yaml")
	var data string = "tasks: nope\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected non-mapping tasks error")
	}

	if !strings.Contains(err.Error(), "tasks must be a mapping") {
		t.Fatalf("expected non-mapping tasks error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":1:8") {
		t.Fatalf("expected non-mapping tasks location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsDuplicateTaskName(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "duplicate-task-name.yaml")
	var data string = "tasks:\n  demo:\n    command: \"echo first\"\n  demo:\n    command: \"echo second\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected duplicate task name error")
	}

	if !strings.Contains(err.Error(), "duplicate task \"demo\"") {
		t.Fatalf("expected duplicate task name error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:3") {
		t.Fatalf("expected duplicate task name location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsBlankTaskName(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "blank-task-name.yaml")
	var data string = "tasks:\n  \"\":\n    command: \"echo nope\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected blank task name error")
	}

	if !strings.Contains(err.Error(), "task name is required") {
		t.Fatalf("expected blank task name error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":2:3") {
		t.Fatalf("expected blank task name location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNonMappingTask(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "non-mapping-task.yaml")
	var data string = "tasks:\n  demo: nope\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected non-mapping task error")
	}

	if !strings.Contains(err.Error(), "task \"demo\" must be a mapping") {
		t.Fatalf("expected non-mapping task error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":2:9") {
		t.Fatalf("expected non-mapping task location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsUnknownTaskField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "unknown-task-field.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build ./...\"\n    depend_on: [\"test\"]\n  test:\n    command: \"go test ./...\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected unknown task field error")
	}

	if !strings.Contains(err.Error(), "unsupported field \"depend_on\" in task \"build\"") {
		t.Fatalf("expected unknown task field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:5") {
		t.Fatalf("expected unknown task field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsDuplicateTaskField(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "duplicate-task-field.yaml")
	var data string = "tasks:\n  demo:\n    command: \"echo first\"\n    command: \"echo second\"\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected duplicate task field error")
	}

	if !strings.Contains(err.Error(), "duplicate field \"command\" in task \"demo\"") {
		t.Fatalf("expected duplicate task field error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:5") {
		t.Fatalf("expected duplicate task field location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsBlankDependencyName(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "blank-dependency.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build ./...\"\n    depends_on: [\"\"]\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected blank dependency name error")
	}

	if !strings.Contains(err.Error(), "task build dependency name is required") {
		t.Fatalf("expected blank dependency name error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:18") {
		t.Fatalf("expected blank dependency name location, got %v", err)
	}
}

func TestLoadWorkflow_RejectsNonListDependsOn(t *testing.T) {
	var tempDir string = t.TempDir()
	var workflowPath string = filepath.Join(tempDir, "non-list-depends-on.yaml")
	var data string = "tasks:\n  build:\n    command: \"go build ./...\"\n    depends_on: test\n"
	var err error = os.WriteFile(workflowPath, []byte(data), 0o644)
	if err != nil {
		t.Fatalf("write workflow file: %v", err)
	}

	_, err = loadWorkflow(workflowPath)
	if err == nil {
		t.Fatal("expected non-list depends_on error")
	}

	if !strings.Contains(err.Error(), "depends_on in task \"build\" must be a list") {
		t.Fatalf("expected non-list depends_on error, got %v", err)
	}
	if !strings.Contains(err.Error(), workflowPath+":4:17") {
		t.Fatalf("expected non-list depends_on location, got %v", err)
	}
}
