# Internals

Talos is small by design. The code is organized around a few clear steps: parse, validate, plan, execute, and summarize.

## Execution Model

Talos treats each workflow as a Directed Acyclic Graph.

- Each task is a node.
- Each `depends_on` entry is an edge.
- A task can run when all of its dependencies have completed successfully.
- Independent tasks can run in parallel.
- Ready tasks are queued by task name before scheduling. This keeps constrained runs, such as `--max-concurrency 1`, predictable while still allowing parallel output to interleave when multiple tasks run at once.

For this workflow:

```text
A -> B
A -> C
B -> D
C -> D
```

Talos runs `A`, then runs `B` and `C` in parallel, then runs `D`.

## Main Components

- `loadWorkflow` parses YAML and applies defaults.
- `validateWorkflow` checks task configuration.
- `validateExecutionOrder` rejects missing dependencies and cycles.
- `buildExecutionPlan` creates the deterministic stage-by-stage dry-run plan.
- `RunWorkflowParallel` schedules ready tasks, tracks completions, and unlocks dependent tasks.
- `VisualizeWorkflow` renders the DAG as Mermaid.

## Failure Behavior

Talos uses fail-fast execution:

- A task only unlocks its dependents after it completes successfully.
- If a task fails, times out, or is canceled, Talos records the first non-cancellation error and cancels the workflow context.
- Running commands receive cancellation through that context.
- Tasks that were ready but not yet running, or tasks blocked behind failed dependencies, are marked as skipped.
- Task output lines are prefixed with the task name so parallel logs stay attributable.
- The final summary shows success, failure, timeout, cancellation, retries, skipped tasks, failure errors, and per-task durations.

Retries happen inside a task attempt loop before the scheduler sees the final result. A task with `retries: 2` can run up to three times. Only the final result is used to decide whether dependents can run.

Timeouts use a task-scoped context derived from the workflow context. If the task context reaches its deadline, the task is reported as timed out and the whole workflow fails.

## Output Modes

Default run output is human-readable. Talos prints task lifecycle events as tasks start, retry, finish, fail, time out, or cancel. Command output is prefixed with the task name so parallel logs remain attributable.

`--quiet` suppresses live lifecycle and command output while keeping the final human summary and final done or failed line.

`--verbose` keeps normal live output and adds task execution context before each task runs: shell, working directory when configured, retry count, timeout, and command. It does not print environment variable values.

`--summary json` suppresses live human output and writes only a machine-readable final summary to stdout. The JSON summary includes overall success, total duration, status counts, and per-task status, attempts, duration, timeout, description, and error details.

## Command Execution

Task commands run through `sh` by default:

```text
sh -c "<task command>"
```

Workflow defaults and task-level configuration can set a different shell executable. Talos still invokes it as:

```text
<shell> -c "<task command>"
```

Talos applies task-specific working directories and environment variables before starting the command. Output emitted by the command is printed with a stable task prefix:

```text
[test] ok ./...
```

## Why This Design

The project favors a small implementation over a large framework. That keeps the behavior easy to inspect while still exercising useful systems concepts:

- graph validation
- deterministic planning
- concurrent scheduling
- context cancellation
- timeout handling
- CLI ergonomics
- testable boundaries
