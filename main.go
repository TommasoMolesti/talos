package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	loadWorkflowFunc      = loadWorkflow
	runWorkflowFunc       = RunWorkflowParallel
	validateWorkflowFunc  = validateWorkflow
	visualizeWorkflowFunc = VisualizeWorkflow
)

// main is the entry point of the Talos CLI application.
//
// It parses CLI arguments and delegates workflow execution
// to the RunWorkflowParallel function.
func main() {
	os.Exit(runCLI(os.Args[1:]))
}

func runCLI(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: talos <command>")
		return 1
	}

	switch args[0] {
	case "run":
		if err := runCmd(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Execution failed:", err)
			return 1
		}
		return 0
	case "validate":
		if err := validateCmd(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Validation failed:", err)
			return 1
		}
		return 0
	case "visualize":
		if err := visualizeCmd(args[1:]); err != nil {
			if err == flag.ErrHelp {
				return 0
			}
			fmt.Fprintln(os.Stderr, "Visualization failed:", err)
			return 1
		}
		return 0
	default:
		fmt.Fprintln(os.Stderr, "Unknown command:", args[0])
		return 1
	}
}

// runCmd handles the "run" command and its flags.
func runCmd(args []string) error {
	// Create a new flag set for the run command
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	// Define flags
	workflowFile := fs.String("file", "talos.yaml", "path to the workflow file")
	dryRun := fs.Bool("dry-run", false, "print the execution plan without running commands")
	maxConcurrency := fs.Int("max-concurrency", 0, "maximum number of concurrent tasks (0 = unlimited)")
	targetTask := fs.String("target", "", "run only the specified task and its dependencies")

	// Parse flags
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	// Load workflow
	wf, err := loadWorkflowFunc(*workflowFile)
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
	opts := RunOptions{
		MaxConcurrency: *maxConcurrency,
		DryRun:         *dryRun,
	}

	return runWorkflowFunc(wf, opts)
}

// validateCmd handles the "validate" command and its flags.
func validateCmd(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	workflowFile := fs.String("file", "talos.yaml", "path to the workflow file")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	wf, err := loadWorkflowFunc(*workflowFile)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	if err := validateWorkflowFunc(wf); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "[talos] Workflow is valid (%d tasks)\n", len(wf.Tasks))
	return nil
}

// visualizeCmd handles the "visualize" command and its flags.
func visualizeCmd(args []string) error {
	fs := flag.NewFlagSet("visualize", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	workflowFile := fs.String("file", "talos.yaml", "path to the workflow file")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return fmt.Errorf("parse flags: %w", err)
	}

	wf, err := loadWorkflowFunc(*workflowFile)
	if err != nil {
		return fmt.Errorf("load workflow: %w", err)
	}

	return visualizeWorkflowFunc(wf)
}
