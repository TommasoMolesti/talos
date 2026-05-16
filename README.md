# Talos

Talos is a lightweight workflow runner for local development. It reads a YAML file, builds a dependency graph, and runs independent tasks in parallel on your machine.

Use it when a project needs repeatable commands such as setup, lint, test, build, migrations, or local service orchestration without adding a server, queue, or external platform.

## Why Talos

In Greek mythology, **Talos** was a giant bronze automaton—the first "robot"—created to protect the island of Crete. Like its namesake, this tool is a self-operating, local-first engine. It doesn't rely on external clouds or complex clusters; it is an autonomous guardian of your workflows, running entirely on your machine to execute tasks with mechanical precision and speed.

## Features

- **Local-first:** runs as a single CLI binary.
- **Dependency-aware:** tasks run only after their dependencies finish.
- **Parallel by default:** independent tasks run concurrently.
- **Safe to preview:** `--dry-run` prints the execution plan before commands run.
- **Practical controls:** per-task `shell`, `cwd`, `env`, `retries`, and `timeout`.
- **Easy to document:** `visualize` exports the workflow DAG as Mermaid.

## Quick Start

Install Talos:

```bash
sh -c "$(curl -sfL https://raw.githubusercontent.com/TommasoMolesti/talos/main/scripts/install.sh)"
```

Or install with Go:

```bash
go install github.com/TommasoMolesti/talos@latest
```

Create a workflow:

```bash
talos init
```

Preview the plan:

```bash
talos run --dry-run
```

Run it:

```bash
talos run
```

## Workflow Example

Talos looks for `talos.yaml` in the current directory by default.

```yaml
defaults:
  shell: "bash"
  timeout: 120

tasks:
  install:
    description: "Install dependencies"
    command: "npm install"

  lint:
    description: "Run lint checks"
    command: "npm run lint"
    depends_on: ["install"]

  test:
    description: "Run tests"
    command: "npm test"
    depends_on: ["install"]

  build:
    description: "Build the app"
    command: "npm run build"
    depends_on: ["lint", "test"]
```

In this workflow, `install` runs first. Then `lint` and `test` run in parallel. `build` runs only after both finish.

## Commands

```bash
talos init                         # create a starter talos.yaml
talos run                          # run talos.yaml
talos run --dry-run                # print the execution plan
talos run --target test            # run one task and its dependencies
talos run --max-concurrency 2      # limit parallel tasks
talos validate                     # check workflow syntax and dependencies
talos visualize                    # print the DAG as Mermaid
talos version                      # print version metadata
```

See [Command Reference](docs/commands.md) for flags and examples.

## Examples

The [examples](examples) directory includes starter workflows for:

- [Go](examples/go.yaml)
- [Node.js](examples/node.yaml)
- [Python](examples/python.yaml)
- [Docker](examples/docker.yaml)
- [Monorepo](examples/monorepo.yaml)

Preview any example:

```bash
talos run --file examples/go.yaml --dry-run
```

## Project Highlights

Talos is intentionally small, but it demonstrates production-oriented engineering choices:

- DAG validation before execution, including missing dependencies and cycles.
- Deterministic execution plans for predictable dry runs and tests.
- Concurrent scheduling with cancellation on failure.
- Per-task shell, retries, timeouts, environment overrides, and working directories.
- User-facing CLI behavior covered by tests.
- Automated release builds for Linux, macOS, and Windows.

## Documentation

- [Workflow Configuration](docs/workflows.md)
- [Command Reference](docs/commands.md)
- [Workflow Patterns](docs/patterns.md)
- [Internals](docs/internals.md)
- [Roadmap](docs/roadmap.md)
- [Release Process](docs/releases.md)

## Development

Run the local checks:

```bash
gofmt -l .
go vet ./...
go test ./...
```

Build the binary:

```bash
go build -o talos .
```

Build with release metadata:

```bash
go build -ldflags "-X main.version=0.2.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o talos .
```
