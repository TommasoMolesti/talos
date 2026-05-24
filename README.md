# Talos

Talos is a small local workflow runner for developers. It reads a YAML file, builds a dependency graph, and runs independent tasks in parallel on your machine.

Use it for repeatable project commands such as setup, lint, test, build, migrations, or local service checks without introducing a CI server or a larger task framework.

## Name

In Greek mythology, **Talos** was a giant bronze automaton—the first "robot"—created to protect the island of Crete. Like its namesake, this tool is a self-operating, local-first engine. It doesn't rely on external clouds or complex clusters; it is an autonomous guardian of your workflows, running entirely on your machine to execute tasks with mechanical precision and speed.

## Why It Exists

Many projects grow a pile of shell scripts, npm scripts, Makefile targets, and README instructions that are easy to run in the wrong order.

Talos keeps that workflow in one file:

- tasks declare their dependencies;
- independent tasks run in parallel;
- `--dry-run` shows the execution plan before anything runs;
- failures, retries, timeouts, and skipped tasks are summarized clearly.

## Quick Start

Install from the latest release:

```bash
sh -c "$(curl -sfL https://raw.githubusercontent.com/TommasoMolesti/talos/main/scripts/install.sh)"
```

Or, if you use Go:

```bash
go install github.com/TommasoMolesti/talos@latest
```

Create a starter workflow:

```bash
talos init
```

Preview and run it:

```bash
talos run --dry-run
talos run
```

## Example

Talos looks for `talos.yaml` in the current directory.

```yaml
tasks:
  install:
    command: "npm install"

  lint:
    command: "npm run lint"
    depends_on: ["install"]

  test:
    command: "npm test"
    depends_on: ["install"]

  build:
    command: "npm run build"
    depends_on: ["lint", "test"]
```

In this workflow, `install` runs first. Then `lint` and `test` run in parallel. `build` runs after both checks pass.

Preview the plan:

```bash
talos run --dry-run
```

## Commands

```bash
talos init
talos run
talos run --dry-run
talos run --target test
talos validate
talos visualize
talos version
```

See [Command Reference](docs/commands.md) for every flag and example.

## What This Project Demonstrates

Talos is intentionally small, but it includes the kind of engineering details that matter in a production CLI:

- YAML schema validation with useful source locations.
- Dependency graph validation, including missing dependencies and cycles.
- Deterministic dry-run plans.
- Concurrent task scheduling with cancellation on failure.
- Live prefixed output so parallel logs remain readable.
- Per-task `cwd`, `shell`, `env`, `retries`, and `timeout`.
- Human and JSON summaries.
- Automated tests and release builds for Linux, macOS, and Windows.

## Examples

The [examples](examples) directory includes workflows for:

- [Go](examples/go.yaml)
- [Node.js](examples/node.yaml)
- [Python](examples/python.yaml)
- [Docker](examples/docker.yaml)
- [Monorepo](examples/monorepo.yaml)
- [Shell configuration](examples/shell.yaml)

Try one with:

```bash
talos run --file examples/go.yaml --dry-run
```

## Documentation

- [Workflow Configuration](docs/workflows.md)
- [Command Reference](docs/commands.md)
- [Workflow Patterns](docs/patterns.md)
- [Internals](docs/internals.md)
- [Release Process](docs/releases.md)
- [Roadmap](docs/roadmap.md)
- [Contributing](CONTRIBUTING.md)

## Development

Run the checks:

```bash
gofmt -l .
sh -n scripts/install.sh
sh -n scripts/uninstall.sh
go vet ./...
go test ./...
```

Build locally:

```bash
go build -o talos .
```
