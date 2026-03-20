package main

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	info = color.New(color.FgCyan).SprintFunc()
	run  = color.New(color.FgYellow).SprintFunc()
	ok   = color.New(color.FgGreen).SprintFunc()
	fail = color.New(color.FgRed).SprintFunc()
)

func PrintStart() {
	fmt.Println(info("[hepa] Starting workflow...\n"))
}

func PrintTaskStart(name string, deps []string) {
	if len(deps) > 0 {
		fmt.Printf("%s %s (depends on: %v)\n", run("▶"), name, deps)
	} else {
		fmt.Printf("%s %s\n", run("▶"), name)
	}
}

func PrintTaskSuccess(name string, duration float64) {
	fmt.Printf("%s %s (%.2fs)\n\n", ok("✔"), name, duration)
}

func PrintTaskFailure(name string, duration float64) {
	fmt.Printf("%s %s (%.2fs)\n\n", fail("✖"), name, duration)
}

func PrintTaskOutputLine(line string) {
	fmt.Println("  " + line)
}

func PrintEnd(total float64) {
	fmt.Printf("%s Done in %.2fs\n", info("[hepa]"), total)
}