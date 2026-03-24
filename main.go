package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	loadWorkflowFunc = loadWorkflow
	runWorkflowFunc  = RunWorkflowParallel
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

	// Run with options
	opts := RunOptions{
		MaxConcurrency: *maxConcurrency,
		DryRun:         *dryRun,
	}

	return runWorkflowFunc(wf, opts)
}
