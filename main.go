package main

import (
	"fmt"
	"os"
	"flag"
)

// main is the entry point of the Talos CLI application.
//
// It parses CLI arguments and delegates workflow execution
// to the RunWorkflowParallel function.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: talos <command>")
		return
	}

	switch os.Args[1] {
	case "run":
		runCmd(os.Args[2:])

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}

// runCmd handles the "run" command and its flags.
func runCmd(args []string) {
	// Create a new flag set for the run command
	fs := flag.NewFlagSet("run", flag.ExitOnError)

	// Define flags
	maxConcurrency := fs.Int("max-concurrency", 0, "maximum number of concurrent tasks (0 = unlimited)")

	// Parse flags
	err := fs.Parse(args)
	if err != nil {
		fmt.Println("Error parsing flags:", err)
		return
	}

	// Load workflow
	wf, err := loadWorkflow("talos.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Run with options
	opts := RunOptions{
		MaxConcurrency: *maxConcurrency,
	}

	err = RunWorkflowParallel(wf, opts)
	if err != nil {
		fmt.Println("Execution failed:", err)
		return
	}
}