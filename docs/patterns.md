# Common Workflow Patterns

These patterns show how to structure local Talos workflows for common developer tasks. Use `talos run --dry-run` before running a new workflow so you can inspect the execution plan.

## Linear Setup, Test, Build

Use a linear chain when each step needs the previous step to succeed.

```yaml
tasks:
  install:
    description: "Install dependencies"
    command: "npm install"

  test:
    description: "Run tests"
    command: "npm test"
    depends_on: ["install"]

  build:
    description: "Build artifacts"
    command: "npm run build"
    depends_on: ["test"]
```

## Parallel Checks

Use shared dependencies to fan out independent tasks. Talos runs `lint` and `test` in parallel after `install` succeeds.

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

## Targeted Runs

Use `--target` when you only want one task plus its required dependencies.

```bash
talos run --target build
```

Given the parallel-checks workflow above, Talos runs `install`, `lint`, `test`, and `build`, but skips unrelated tasks.

## Shared Defaults

Use workflow defaults for settings that most tasks share. Task-level values override defaults.

```yaml
defaults:
  cwd: "."
  env:
    APP_ENV: "development"
  retries: 1
  timeout: 60

tasks:
  test:
    command: "go test ./..."

  migrate:
    command: "go run ./cmd/migrate"
    cwd: "./backend"
    timeout: 30
```

## Task-Local Directories And Environment

Use `cwd` and `env` for commands that need to run from a package directory or need task-specific configuration.

```yaml
tasks:
  api-test:
    command: "go test ./..."
    cwd: "./services/api"
    env:
      DATABASE_URL: "postgres://localhost:5432/app"

  web-test:
    command: "npm test"
    cwd: "./apps/web"
```

## Flaky Or Slow Commands

Use `retries` for transient failures and `timeout` for commands that should not run forever.

```yaml
tasks:
  integration-test:
    command: "npm run test:integration"
    retries: 2
    timeout: 120
```

## Local Services

Use dependencies to start services before migrations or tests. Keep teardown commands explicit so users can choose when to run them.

```yaml
tasks:
  db:
    description: "Start local database"
    command: "docker compose up -d db"

  migrate:
    command: "npm run migrate"
    depends_on: ["db"]

  test:
    command: "npm test"
    depends_on: ["migrate"]

  down:
    description: "Stop local services"
    command: "docker compose down"
```

Run the main path with:

```bash
talos run --target test
```

Then stop services when you are done:

```bash
talos run --target down
```

## Visual Documentation

Use `visualize` to include workflow graphs in documentation or pull requests.

```bash
talos visualize --file talos.yaml
```
