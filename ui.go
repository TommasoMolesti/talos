package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

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

// PrintDryRun prints the planned workflow execution without running commands.
func PrintDryRun(plan [][]string, wf *Workflow) {
	fmt.Println(info("[talos] Dry run\n"))
	for i, stage := range plan {
		fmt.Printf("%s Stage %d: %s\n", run("▶"), i+1, strings.Join(stage, ", "))
		for _, name := range stage {
			fmt.Printf("  - %s: %s\n", name, wf.Tasks[name].Command)
		}
		fmt.Println()
	}
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

// PrintTaskRetry prints a retry message before a new attempt begins.
func PrintTaskRetry(name string, attempt int, maxAttempts int, err error) {
	fmt.Printf("%s %s retry %d/%d after error: %v\n", run("↻"), name, attempt, maxAttempts, err)
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

// PrintTaskTimeout prints a timeout message for a task,
// including its execution duration and timeout limit.
func PrintTaskTimeout(name string, duration float64, timeout time.Duration) {
	fmt.Printf("%s %s (%.2fs, timeout %s)\n\n", fail("⌛"), name, duration, timeout)
}

// PrintEnd prints the final message after workflow execution completes,
// including the total execution time.
func PrintEnd(total float64, success bool) {
	if success {
		fmt.Printf("%s Done in %.2fs\n", info("[talos]"), total)
		return
	}
	fmt.Printf("%s Failed in %.2fs\n", fail("[talos]"), total)
}

// PrintSummary prints a final summary of task outcomes for the workflow run.
func PrintSummary(summary *executionSummary) {
	counts := make(map[taskStatus]int)
	var retried []string
	var timedOut []string
	var canceled []string
	var skippedTasks []string

	for name, task := range summary.Tasks {
		counts[task.Status]++
		if task.Attempts > 1 {
			retried = append(retried, fmt.Sprintf("%s (%d retries)", name, task.Attempts-1))
		}
		if task.Status == taskStatusTimedOut {
			timedOut = append(timedOut, fmt.Sprintf("%s (%s)", name, task.Timeout))
		}
		if task.Status == taskStatusCanceled {
			canceled = append(canceled, name)
		}
		if task.Status == taskStatusSkipped {
			skippedTasks = append(skippedTasks, name)
		}
	}

	sort.Strings(retried)
	sort.Strings(timedOut)
	sort.Strings(canceled)
	sort.Strings(skippedTasks)

	fmt.Println()
	fmt.Println(info("[talos] Summary"))
	fmt.Printf(
		"  total=%d success=%d failed=%d timed_out=%d canceled=%d skipped=%d\n",
		len(summary.Tasks),
		counts[taskStatusSuccess],
		counts[taskStatusFailed],
		counts[taskStatusTimedOut],
		counts[taskStatusCanceled],
		counts[taskStatusSkipped],
	)
	if len(retried) > 0 {
		fmt.Printf("  retries: %s\n", strings.Join(retried, ", "))
	}
	if len(timedOut) > 0 {
		fmt.Printf("  timeouts: %s\n", strings.Join(timedOut, ", "))
	}
	if len(canceled) > 0 {
		fmt.Printf("  canceled: %s\n", strings.Join(canceled, ", "))
	}
	if len(skippedTasks) > 0 {
		fmt.Printf("  skipped: %s\n", strings.Join(skippedTasks, ", "))
	}
}
