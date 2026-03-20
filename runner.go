package main

import (
	"os/exec"
	"strings"
)

func runTask(task *Task) error {
	cmd := exec.Command("sh", "-c", task.Command)

	output, err := cmd.CombinedOutput()

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			PrintTaskOutputLine(line)
		}
	}

	return err
}