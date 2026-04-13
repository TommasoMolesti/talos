package main

import (
	"fmt"
	"sort"
	"strings"
)

// VisualizeWorkflow prints the workflow DAG in Mermaid format.
func VisualizeWorkflow(wf *Workflow) error {
	var err error = validateExecutionOrder(wf)
	if err != nil {
		return err
	}

	fmt.Print(BuildMermaidGraph(wf))
	return nil
}

// BuildMermaidGraph renders the workflow DAG as a Mermaid graph.
func BuildMermaidGraph(wf *Workflow) string {
	var names []string = sortedTaskNames(wf)

	var b strings.Builder
	b.WriteString("graph TD\n")

	for _, name := range names {
		fmt.Fprintf(&b, "    %s[\"%s\"]\n", mermaidID(name), name)
	}

	var edges []string
	for _, name := range names {
		var task *Task = wf.Tasks[name]
		for _, dep := range task.DependsOn {
			edges = append(edges, fmt.Sprintf("    %s --> %s\n", mermaidID(dep), mermaidID(name)))
		}
	}
	sort.Strings(edges)
	for _, edge := range edges {
		b.WriteString(edge)
	}

	return b.String()
}

func sortedTaskNames(wf *Workflow) []string {
	var names []string = make([]string, 0, len(wf.Tasks))
	for name := range wf.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func mermaidID(name string) string {
	var b strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}
