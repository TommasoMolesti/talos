package main

type Task struct {
	Name      string   `yaml:"-"`
	Command   string   `yaml:"command"`
	DependsOn []string `yaml:"depends_on"`
}

type Workflow struct {
	Tasks map[string]*Task `yaml:"tasks"`
}