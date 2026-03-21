package main

import (
	"fmt"
	"os"
)

// main is the entry point of the Talos CLI application.
//
// It parses CLI arguments and delegates workflow execution
// to the RunWorkflowParallel function. It handles user-facing errors
// and command routing.
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: talos <command>")
		return
	}

	switch os.Args[1] {
	case "run":
		wf, err := loadWorkflow("talos.yaml")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		err = RunWorkflowParallel(wf)
		if err != nil {
			fmt.Println("Execution failed:", err)
			return
		}

	default:
		fmt.Println("Unknown command:", os.Args[1])
	}
}