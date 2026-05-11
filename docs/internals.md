# Internals

Talos is small by design. The code is organized around a few clear steps: parse, validate, plan, execute, and summarize.

## Execution Model

Talos treats each workflow as a Directed Acyclic Graph.

- Each task is a node.
- Each `depends_on` entry is an edge.
- A task can run when all of its dependencies have completed successfully.
- Independent tasks can run in parallel.

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
- The final summary shows success, failure, timeout, cancellation, retries, and skipped tasks.

Retries happen inside a task attempt loop before the scheduler sees the final result. A task with `retries: 2` can run up to three times. Only the final result is used to decide whether dependents can run.

Timeouts use a task-scoped context derived from the workflow context. If the task context reaches its deadline, the task is reported as timed out and the whole workflow fails.

## Command Execution

Task commands run through the system shell:

```text
sh -c "<task command>"
```

Talos applies task-specific working directories and environment variables before starting the command.

## Why This Design

The project favors a small implementation over a large framework. That keeps the behavior easy to inspect while still exercising useful systems concepts:

- graph validation
- deterministic planning
- concurrent scheduling
- context cancellation
- timeout handling
- CLI ergonomics
- testable boundaries
