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

- If a task fails, Talos stops scheduling new tasks.
- Running commands receive cancellation.
- Downstream tasks that can no longer run are marked as skipped.
- The final summary shows success, failure, timeout, cancellation, retries, and skipped tasks.

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
