# Talos

**Talos** is a lightweight, local-first workflow engine designed to execute task pipelines directly on your machine.

It allows you to define workflows as a set of tasks with dependencies, forming a **Directed Acyclic Graph (DAG)**. Talos resolves execution order automatically and runs tasks efficiently, executing independent tasks in parallel.

The goal of Talos is to provide a simple, portable alternative to heavier orchestration tools without requiring any external infrastructure.

---

## Why Talos?

In Greek mythology, **Talos** was a giant bronze automaton—the first "robot"—created to protect the island of Crete. Like its namesake, this tool is a self-operating, local-first engine. It doesn't rely on external clouds or complex clusters; it is an autonomous guardian of your workflows, running entirely on your machine to execute tasks with mechanical precision and speed.

## Features

- **Local-first execution:** No servers, no clusters, just your machine.
- **Zero setup:** A single binary is all you need.
- **YAML-based configuration:** Simple, readable workflow definitions.
- **Dependency-aware:** Smart execution based on task relationships.
- **Parallel execution:** Runs independent tasks concurrently to save time.
- **Dry-run mode:** Prints the execution plan without touching your system.
- **Best-effort fail-fast behavior:** Stops scheduling new tasks and cancels running commands when a task fails.
- **Clean CLI output:** Real-time updates on your pipeline's progress.

## Quick Start

Build the binary:

```bash
go build -o talos .
```

Run the sample workflow in this repository:

```bash
./talos run
```

Run a workflow from a custom path:

```bash
./talos run --file ./workflows/dev.yaml
```

Preview the execution plan without running commands:

```bash
./talos run --dry-run
```

Limit parallelism:

```bash
./talos run --max-concurrency 2
```

### Create a workflow

```yaml
tasks:
  install:
    command: "npm install"

  db:
    command: "docker-compose up -d"

  migrate:
    command: "npm run migrate"
    depends_on: ["db"]

  dev:
    command: "npm run dev"
    depends_on: ["install", "migrate"]
```

By default, Talos looks for `talos.yaml` in the current directory, but you can override that with `--file`.
Use `--dry-run` to inspect the execution stages and commands before you run a workflow for real.

## How It Works

Talos parses your workflow into a Directed Acyclic Graph (DAG):

- Tasks run only after their dependencies are completed
- Independent tasks are executed in parallel
- Execution is event-driven: when a task finishes, it unlocks dependent tasks

Example:

A → B, C → D

Execution:

1. A runs first
2. B and C run in parallel
3. D runs after both B and C complete

## Architecture Notes

Talos is intentionally small, but it still shows a few useful Go design ideas:

- `loadWorkflow` parses YAML and normalizes task names.
- `validateExecutionOrder` checks for missing dependencies and cycles before any command starts.
- `RunWorkflowParallel` uses dependency counts plus a result loop to schedule ready tasks and unlock dependents as tasks finish.
- Command execution uses `exec.CommandContext`, so a failing task can cancel other running commands.

This keeps the code easy to follow while still demonstrating concurrency, graph traversal, and CLI design.

## Testing

```bash
go test ./...
```

## Roadmap

- visualization
- retries and timeouts
