package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func loadWorkflow(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	err = yaml.Unmarshal(data, &wf)
	if err != nil {
		return nil, err
	}

	for name, task := range wf.Tasks {
		task.Name = name
	}

	return &wf, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hepa <command>")
		return
	}

	switch os.Args[1] {
		case "run":
			wf, err := loadWorkflow("hepa.yaml")
			if err != nil {
				fmt.Println("Error:", err)
				return
			}

			PrintStart()

			totalStart := time.Now()

			for _, task := range wf.Tasks {
				start := time.Now()

				PrintTaskStart(task.Name, task.DependsOn)

				err := runTask(task)

				duration := time.Since(start).Seconds()

				if err != nil {
					PrintTaskFailure(task.Name, duration)
					return
				}

				PrintTaskSuccess(task.Name, duration)
			}

			PrintEnd(time.Since(totalStart).Seconds())

		default:
			fmt.Println("Unknown command:", os.Args[1])
	}
}