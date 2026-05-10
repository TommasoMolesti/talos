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

## Working Directory

Use `cwd` when a command must run from a specific directory.

```yaml
tasks:
  api-test:
    command: "go test ./..."
    cwd: "./services/api"
```

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

## Defaults

Use `defaults` for values shared by most tasks.

```yaml
defaults:
  cwd: "."
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
