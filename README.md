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
- **Per-task retries:** Retry flaky commands before failing the workflow.
- **Per-task timeouts:** Cancel tasks that run longer than expected.
- **Per-task working directories and environment:** Run commands with task-specific `cwd` and `env` settings.
- **Visualization:** Export the workflow DAG as a Mermaid graph for docs and demos.
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

Validate a workflow file without running any commands:

```bash
./talos validate
```

Render the workflow as a Mermaid DAG:

```bash
./talos visualize
```

Limit parallelism:

```bash
./talos run --max-concurrency 2
```

Run only one task and the dependencies needed to reach it:

```bash
./talos run --target test
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
    cwd: "./backend"
    env:
      DATABASE_URL: "postgres://localhost:5432/app"
    retries: 2
    depends_on: ["db"]
    timeout: 30

  dev:
    command: "npm run dev"
    depends_on: ["install", "migrate"]
```

By default, Talos looks for `talos.yaml` in the current directory, but you can override that with `--file`.
Use `--dry-run` to inspect the execution stages and commands before you run a workflow for real.
Use `--target <task>` to run just one part of a workflow, including only the dependencies required for that task.
Use `validate` to verify YAML parsing, task settings, and DAG correctness without starting any commands.
Use `visualize` to export the workflow graph in Mermaid format for README snippets or architecture docs.
Use `cwd` and `env` on a task when commands need to run from a specific directory or with task-local environment variables.
Use `retries` on a task to retry transient failures before Talos gives up.
Use `timeout` on a task to fail fast when a command exceeds its expected runtime. Timeout values are expressed as seconds.
Use `talos -h` or `talos <command> -h` to see command-specific examples and flag guidance.

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
- `buildExecutionPlan` produces the deterministic stage-by-stage plan shown by `--dry-run`, grouping tasks that can run in parallel.
- `RunWorkflowParallel` uses dependency counts plus a result loop to schedule ready tasks and unlock dependents as tasks finish.
- Command execution uses `exec.CommandContext`, so a failing or timed-out task can cancel other running commands, while transient errors can be retried per task.

This keeps the code easy to follow while still demonstrating concurrency, graph traversal, and CLI design.

## Testing

```bash
go test ./...
```

## Roadmap

- shell tab completion for commands and flags
- final execution summary with retries, timeouts, and skips
