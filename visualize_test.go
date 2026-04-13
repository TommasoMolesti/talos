package main

import (
	"strings"
	"testing"
)

func TestBuildMermaidGraph_SortsNodesAndEdges(t *testing.T) {
	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"deploy": {Name: "deploy", DependsOn: []string{"test", "build"}},
			"build":  {Name: "build"},
			"test":   {Name: "test", DependsOn: []string{"build"}},
		},
	}

	var got string = BuildMermaidGraph(wf)
	var want string = strings.Join([]string{
		"graph TD",
		"    build[\"build\"]",
		"    deploy[\"deploy\"]",
		"    test[\"test\"]",
		"    build --> deploy",
		"    build --> test",
		"    test --> deploy",
		"",
	}, "\n")

	if got != want {
		t.Fatalf("unexpected mermaid graph:\n%s", got)
	}
}

func TestBuildMermaidGraph_SanitizesNodeIDs(t *testing.T) {
	var wf *Workflow = &Workflow{
		Tasks: map[string]*Task{
			"lint-check": {Name: "lint-check"},
		},
	}

	var got string = BuildMermaidGraph(wf)
	if !strings.Contains(got, "lint_check[\"lint-check\"]") {
		t.Fatalf("expected sanitized node id, got %q", got)
	}
}
