package main

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	info = color.New(color.FgCyan).SprintFunc()
	run  = color.New(color.FgYellow).SprintFunc()
	ok   = color.New(color.FgGreen).SprintFunc()
	skip = color.New(color.FgHiBlack).SprintFunc()
	fail = color.New(color.FgRed).SprintFunc()
)

// PrintStart prints the initial message indicating that workflow execution has started.
func PrintStart() {
	fmt.Println(info("[talos] Starting workflow...\n"))
}

// PrintTaskStart prints the header for a task before execution.
//
// It includes the task name and optionally its dependencies.
func PrintTaskStart(name string, deps []string) {
	if len(deps) > 0 {
		fmt.Printf("%s %s (depends on: %v)\n", run("▶"), name, deps)
	} else {
		fmt.Printf("%s %s\n", run("▶"), name)
	}
}

// PrintTaskOutputLine prints a single line of task output,
// formatted with indentation to distinguish it from system logs.
func PrintTaskOutputLine(line string) {
	fmt.Println("  " + line)
}

// PrintTaskSuccess prints a success message for a completed task,
// including its execution duration in seconds.
func PrintTaskSuccess(name string, duration float64) {
	fmt.Printf("%s %s (%.2fs)\n\n", ok("✔"), name, duration)
}

// PrintTaskFailure prints a failure message for a task,
// including its execution duration in seconds.
func PrintTaskFailure(name string, duration float64) {
	fmt.Printf("%s %s (%.2fs)\n\n", fail("✖"), name, duration)
}

// PrintTaskCanceled prints a cancellation message for a task,
// including its execution duration in seconds.
func PrintTaskCanceled(name string, duration float64) {
	fmt.Printf("%s %s (%.2fs)\n\n", skip("◌"), name, duration)
}

// PrintEnd prints the final message after workflow execution completes,
// including the total execution time.
func PrintEnd(total float64) {
	fmt.Printf("%s Done in %.2fs\n", info("[talos]"), total)
}
