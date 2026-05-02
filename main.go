package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var (
	loadWorkflowFunc      = loadWorkflow
	runWorkflowFunc       = RunWorkflowParallel
	validateWorkflowFunc  = validateWorkflow
	visualizeWorkflowFunc = VisualizeWorkflow
)

const starterWorkflow = `tasks:
  setup:
    description: "Prepare the project"
    command: "echo preparing project"

  test:
    description: "Run the test suite"
    command: "go test ./..."
    depends_on: ["setup"]
`

// main is the entry point of the Talos CLI application.
//
// It parses CLI arguments and delegates workflow execution
// to the RunWorkflowParallel function.
func main() {
	os.Exit(runCLI(os.Args[1:]))
}

// runCLI dispatches parsed command-line arguments to the matching command.
func runCLI(args []string) int {
	if len(args) < 1 {
		printRootUsage(os.Stderr)
		return 1
	}

	switch args[0] {
	case "-h", "--help", "help":
		printRootUsage(os.Stdout)
		return 0
	case "init":
		var err error = initCmd(args[1:])
		if err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Init failed:", err)
			return 1
		}
		return 0
	case "run":
		var err error = runCmd(args[1:])
		if err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Execution failed:", err)
			return 1
		}
		return 0
	case "validate":
		var err error = validateCmd(args[1:])
		if err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Validation failed:", err)
			return 1
		}
		return 0
	case "visualize":
		var err error = visualizeCmd(args[1:])
		if err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Visualization failed:", err)
			return 1
		}
		return 0
	case "version":
		printVersion()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", args[0])
		fmt.Fprintln(os.Stderr)
		printRootUsage(os.Stderr)
		return 1
	}
}

// initCmd handles the "init" command and its flags.
func initCmd(args []string) error {
	var fs *flag.FlagSet = flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: talos init [flags]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Create a starter workflow file.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  talos init")
		fmt.Fprintln(os.Stderr, "  talos init --file ./workflows/dev.yaml")
		fmt.Fprintln(os.Stderr, "  talos init --force")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	var workflowFile *string = fs.String("file", "talos.yaml", "path to write the starter workflow")
	var force *bool = fs.Bool("force", false, "overwrite an existing workflow file")

	var err error = fs.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	var flag int = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if *force {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}

	var dir string = filepath.Dir(*workflowFile)
	if dir != "." {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return fmt.Errorf("create workflow directory: %w", err)
		}
	}

	var file *os.File
	file, err = os.OpenFile(*workflowFile, flag, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists; use --force to overwrite", *workflowFile)
		}
		return fmt.Errorf("create workflow file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(starterWorkflow)
	if err != nil {
		return fmt.Errorf("write workflow file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "[talos] Created starter workflow at %s\n", *workflowFile)
	return nil
}

// runCmd handles the "run" command and its flags.
func runCmd(args []string) error {
	// Create a new flag set for the run command
	var fs *flag.FlagSet = flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: talos run [flags]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Execute a workflow from a YAML file.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  talos run")
		fmt.Fprintln(os.Stderr, "  talos run --file ./workflows/dev.yaml")
		fmt.Fprintln(os.Stderr, "  talos run --dry-run --target test")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	// Define flags
	var workflowFile *string = fs.String("file", "talos.yaml", "path to the workflow file")
	var dryRun *bool = fs.Bool("dry-run", false, "print the execution plan without running commands")
	var maxConcurrency *int = fs.Int("max-concurrency", 0, "maximum number of concurrent tasks (0 = unlimited)")
	var targetTask *string = fs.String("target", "", "run only the specified task and its dependencies")

	// Parse flags
	var err error = fs.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	// Load workflow
	var wf *Workflow
	wf, err = loadWorkflowFunc(*workflowFile)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	if *targetTask != "" {
		wf, err = workflowForTarget(wf, *targetTask)
		if err != nil {
			return err
		}
	}

	// Run with options
	var opts RunOptions = RunOptions{
		MaxConcurrency: *maxConcurrency,
		DryRun:         *dryRun,
	}

	return runWorkflowFunc(wf, opts)
}

// validateCmd handles the "validate" command and its flags.
func validateCmd(args []string) error {
	var fs *flag.FlagSet = flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: talos validate [flags]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Validate a workflow file without running commands.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  talos validate")
		fmt.Fprintln(os.Stderr, "  talos validate --file ./workflows/dev.yaml")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	var workflowFile *string = fs.String("file", "talos.yaml", "path to the workflow file")

	var err error = fs.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	var wf *Workflow
	wf, err = loadWorkflowFunc(*workflowFile)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	err = validateWorkflowFunc(wf)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "[talos] Workflow is valid (%d tasks)\n", len(wf.Tasks))
	return nil
}

// visualizeCmd handles the "visualize" command and its flags.
func visualizeCmd(args []string) error {
	var fs *flag.FlagSet = flag.NewFlagSet("visualize", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: talos visualize [flags]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Render the workflow DAG as a Mermaid graph.")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Examples:")
		fmt.Fprintln(os.Stderr, "  talos visualize")
		fmt.Fprintln(os.Stderr, "  talos visualize --file ./workflows/dev.yaml")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	var workflowFile *string = fs.String("file", "talos.yaml", "path to the workflow file")

	var err error = fs.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	var wf *Workflow
	wf, err = loadWorkflowFunc(*workflowFile)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	return visualizeWorkflowFunc(wf)
}

// printVersion writes the binary version and build metadata.
func printVersion() {
	fmt.Fprintf(os.Stdout, "talos %s\n", version)
	fmt.Fprintf(os.Stdout, "commit: %s\n", commit)
	fmt.Fprintf(os.Stdout, "built: %s\n", date)
}

// printRootUsage writes top-level command help to the given stream.
func printRootUsage(w *os.File) {
	fmt.Fprintln(w, "Talos executes local task workflows described as dependency-aware DAGs.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  talos <command> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  init       Create a starter workflow file")
	fmt.Fprintln(w, "  run        Execute a workflow")
	fmt.Fprintln(w, "  validate   Check workflow syntax and dependency correctness")
	fmt.Fprintln(w, "  visualize  Print the workflow DAG as Mermaid")
	fmt.Fprintln(w, "  version    Print version and build metadata")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  talos init")
	fmt.Fprintln(w, "  talos run --dry-run")
	fmt.Fprintln(w, "  talos run --target build")
	fmt.Fprintln(w, "  talos validate --file ./workflows/dev.yaml")
	fmt.Fprintln(w, "  talos visualize")
	fmt.Fprintln(w, "  talos version")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use \"talos <command> -h\" for command-specific help.")
}
