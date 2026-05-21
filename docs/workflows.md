# Workflow Configuration

Talos workflows are YAML files. By default, Talos reads `talos.yaml` from the current directory.

Each workflow has a `tasks` map. Each task defines one shell command and can depend on other tasks.

## Minimal Workflow

```yaml
tasks:
  test:
    command: "go test ./..."
```

Run it with:

```bash
talos run
```

## Tasks

```yaml
tasks:
  build:
    description: "Build the binary"
    command: "go build -o talos ."
```

Task names must be unique. The task name is used by `depends_on`, `--target`, dry-run output, and execution summaries.

Every task must define a non-empty `command`. Talos validates this before dry-run or execution.

## Supported Fields

Top-level workflow fields:

| Field | Required | Description |
| --- | --- | --- |
| `tasks` | Yes | Map of task names to task definitions. |
| `defaults` | No | Shared task settings applied before task-level overrides. |

Task fields:

| Field | Required | Description |
| --- | --- | --- |
| `command` | Yes | Shell command to run. |
| `description` | No | Human-readable task label for dry-run and summary output. |
| `depends_on` | No | List of task names that must succeed first. |
| `cwd` | No | Workflow-relative or absolute working directory. |
| `shell` | No | Shell executable used as `<shell> -c "<command>"`. |
| `env` | No | Environment variable overrides. |
| `retries` | No | Number of retry attempts after the first failure. |
| `timeout` | No | Timeout in seconds. |

Default fields:

| Field | Description |
| --- | --- |
| `cwd` | Shared working directory. |
| `shell` | Shared shell executable. |
| `env` | Shared environment variables. |
| `retries` | Shared retry count. |
| `timeout` | Shared timeout in seconds. |

## Compatibility

For `v1.x`, Talos treats the fields listed above as the stable workflow schema. Patch and minor releases may fix bugs, improve diagnostics, and add compatible behavior, but they should not remove these fields or change their meaning.

Talos rejects unsupported top-level, `defaults`, and task fields with a file, line, and column error. This keeps typos such as `depend_on` from silently producing a different execution plan.

Workflow files should not rely on undocumented fields, task map ordering, exact human-readable colors or symbols, or shell behavior that is specific to a machine unless the workflow explicitly configures that shell.

## Dependencies

Use `depends_on` when a task must wait for another task.

```yaml
tasks:
  test:
    command: "go test ./..."

  build:
    command: "go build -o talos ."
    depends_on: ["test"]
```

Talos validates dependencies before running commands. If a dependency is missing or the workflow contains a cycle, execution stops with an error.

Each task should list a dependency only once. Duplicate `depends_on` entries are rejected during validation because they describe the same graph edge twice.

## Parallel Execution

Tasks without a dependency relationship can run at the same time.

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

Here, `lint` and `test` run in parallel after `install`.

## Descriptions

Use `description` to make output easier to scan.

```yaml
tasks:
  migrate:
    description: "Run database migrations"
    command: "npm run migrate"
```

Descriptions appear in dry-run and summary output.

During execution, command output is prefixed with the task name so concurrent runs stay readable:

```text
[migrate] applied 3 migrations
```

## Working Directory

Use `cwd` when a command must run from a specific directory.

```yaml
tasks:
  api-test:
    command: "go test ./..."
    cwd: "./services/api"
```

## Shell

Talos runs commands through `sh` by default:

```text
sh -c "<command>"
```

Use `shell` when a workflow or task needs a different shell executable.

```yaml
defaults:
  shell: "bash"

tasks:
  test:
    command: "go test ./..."

  release:
    command: "set -euo pipefail; ./scripts/release.sh"
    shell: "zsh"
```

Task-level `shell` values override the workflow default. Talos passes the command to the configured shell as `-c "<command>"`; it does not translate shell syntax, quoting, environment expansion, path behavior, or built-ins across shells or operating systems.

Use a shell name that exists on the machines where the workflow will run, such as `bash`, `zsh`, or an absolute path to a shell executable.

Dry-run output includes non-default shell selections so you can see which shell each configured task will use before commands run.

Validation and dry-run do not check whether the configured shell executable exists. A missing shell is reported when Talos tries to run the task.

See [the shell example](../examples/shell.yaml) for workflow-level shell configuration and a task-level override.

## Environment Variables

Use `env` for task-specific environment variables.

```yaml
tasks:
  integration-test:
    command: "npm run test:integration"
    env:
      DATABASE_URL: "postgres://localhost:5432/app"
```

Talos keeps the current process environment and applies the task values on top.

## Retries

Use `retries` for commands that can fail temporarily.

```yaml
tasks:
  integration-test:
    command: "npm run test:integration"
    retries: 2
```

`retries: 2` means Talos can try the task up to three times: the first attempt plus two retries.

## Timeouts

Use `timeout` to stop a task that runs too long.

```yaml
tasks:
  smoke-test:
    command: "npm run smoke"
    timeout: 60
```

Timeout values are seconds.

## Failure Semantics

Talos uses fail-fast execution. When any task fails, times out, or is canceled, the workflow fails.

During a failed run:

- Tasks whose dependencies completed successfully may already be running.
- Running tasks receive cancellation.
- Talos stops scheduling new tasks after the first failure.
- Tasks that never started are marked as skipped in the final summary.
- Dependents of a failed, timed-out, canceled, or skipped task do not run.

Retries are handled before a task is considered failed. For example, `retries: 2` allows one initial attempt and two retry attempts. If the final attempt fails, the workflow enters fail-fast cancellation.

Timeouts stop the current task and fail the workflow immediately. Timed-out tasks are reported separately from ordinary command failures.

When a task fails, the final summary includes the failed task name and the command error returned by the shell or process.

Use `talos run --quiet` to suppress live task output while keeping the final summary.

Use `talos run --verbose` to print each task's shell, working directory, retry count, timeout, and command before it runs. Verbose output does not print environment variable values.

Use `talos run --summary json` when scripts need a machine-readable final summary. JSON summary mode suppresses live task output and prints only the summary JSON to stdout.

## Defaults

Use `defaults` for values shared by most tasks.

```yaml
defaults:
  cwd: "."
  shell: "bash"
  retries: 1
  timeout: 120
  env:
    APP_ENV: "development"

tasks:
  test:
    command: "go test ./..."

  migrate:
    command: "npm run migrate"
    cwd: "./backend"
    timeout: 30
```

Task-level values override defaults.

## Full Example

```yaml
defaults:
  cwd: "."
  timeout: 120

tasks:
  fmt:
    description: "Check Go formatting"
    command: "test -z \"$(gofmt -l .)\""

  vet:
    description: "Run Go vet"
    command: "go vet ./..."
    depends_on: ["fmt"]

  test:
    description: "Run Go tests"
    command: "go test ./..."
    depends_on: ["vet"]

  build:
    description: "Build the binary"
    command: "go build -o bin/app ."
    depends_on: ["test"]
```
