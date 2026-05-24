# v1 Validation Plan

This checklist records the real-project testing required before Talos can become a `v1.0.0` release candidate.

The bundled examples prove that the documented schema parses and dry-runs. This plan is for testing Talos against existing projects with real commands, toolchains, working directories, and failure modes.

## Build Under Test

Before testing real projects, build one local binary and use the same binary for every validation run:

```bash
go build -o /tmp/talos-v1-rc .
/tmp/talos-v1-rc version
```

If the binary uses release-candidate metadata, record the version, commit, and build timestamp from `talos version`.

## Project Matrix

Test at least one project in each category:

| Category | Expected coverage |
| --- | --- |
| Go | Formatting, vetting, tests, and build tasks. |
| Node.js | Install, lint, test, and build tasks. |
| Python | Virtual environment or dependency setup, lint or type checks, tests, and packaging or build tasks. |
| Docker | Compose or container tasks, service startup, tests, and teardown. |
| Monorepo | Multiple `cwd` values, shared defaults, task-local environment, fan-out/fan-in dependencies, retries, and timeouts. |

Use existing project workflows when possible. If a project does not already have `talos.yaml`, create one outside the Talos repository and keep a copy of the workflow used for validation.

## Per-Project Checks

For each project:

```bash
/tmp/talos-v1-rc validate --file talos.yaml
/tmp/talos-v1-rc run --file talos.yaml --dry-run
/tmp/talos-v1-rc visualize --file talos.yaml
```

Then run the project's normal validation target:

```bash
/tmp/talos-v1-rc run --file talos.yaml --target <main-check-target>
```

If the workflow has safe full-run semantics, also run:

```bash
/tmp/talos-v1-rc run --file talos.yaml
```

Use `--quiet`, `--verbose`, and `--summary json` on at least one real workflow so output modes are checked outside unit tests:

```bash
/tmp/talos-v1-rc run --file talos.yaml --target <main-check-target> --quiet
/tmp/talos-v1-rc run --file talos.yaml --target <main-check-target> --verbose
/tmp/talos-v1-rc run --file talos.yaml --target <main-check-target> --summary json
```

## Failure Checks

At least one real-project validation should confirm failure behavior:

- A missing dependency is rejected by `validate`.
- An unsupported or duplicate workflow key is rejected with a source location.
- A failing task stops new scheduling and marks blocked dependents as skipped.
- A timed-out task is reported as `timed_out`.

Do these in a temporary branch or throwaway workflow file so project configuration is not damaged.

## Evidence Log

Record the results before creating a release candidate.

| Category | Project | Commit | Workflow file | Commands run | Result | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Go |  |  |  |  |  |  |
| Node.js |  |  |  |  |  |  |
| Python |  |  |  |  |  |  |
| Docker |  |  |  |  |  |  |
| Monorepo |  |  |  |  |  |  |

The release candidate is ready only after each category has a passing result or a documented, intentional exclusion.
