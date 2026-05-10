# Roadmap

Talos is a local-first workflow runner for developers who want dependency-aware task execution without external infrastructure.

The roadmap is intentionally conservative. Talos should become more reliable, easier to adopt, and clearer under failure before it grows into heavier orchestration features.

## Current Focus

Talos already supports:

- YAML workflow files.
- Dependency validation.
- Parallel task execution.
- Dry-run execution plans.
- Targeted runs with `--target`.
- Per-task `cwd`, `env`, `retries`, and `timeout`.
- Mermaid DAG visualization.
- Release binaries for multiple platforms.

The next work should make those foundations feel stable enough for real projects.

## Guiding Principles

- Keep Talos local-first and dependency-light.
- Prefer explicit workflow behavior over hidden magic.
- Make failure states easy to understand.
- Keep the default CLI output human-readable.
- Add automation features without making interactive usage noisy.

## Next Implementation Slices

These are the best next PR-sized changes. Each one should be small enough to review on its own.

| Priority | Change | Why It Matters |
| --- | --- | --- |
| 1 | Document failure semantics | Users should know exactly what happens when a task fails, times out, or is skipped. |
| 2 | Add configurable shell support | Some workflows need `bash`, `zsh`, or platform-specific shell behavior. |
| 3 | Add Mermaid file output | `talos visualize --output workflow.md` makes docs and pull requests easier. |
| 4 | Add compact output mode | Long workflows need less noisy progress output. |
| 5 | Add JSON summary output | Scripts and CI can consume results without parsing terminal text. |

## Milestone 1: Execution Semantics

Goal: make task behavior explicit, predictable, and easy to reason about.

Planned work:

- Add configurable shell support.
- Add `continue_on_error` for non-blocking tasks.
- Add CLI task skipping, such as skipping known slow or optional tasks.
- Document exact behavior for failed, skipped, canceled, and timed-out tasks.
- Add tests for mixed failure scenarios across parallel branches.

Suggested implementation order:

1. Write failure behavior docs for the current implementation.
2. Add tests that lock in current fail-fast behavior.
3. Add workflow-level and task-level shell configuration.
4. Add `continue_on_error` only after failure semantics are documented.
5. Add skip behavior after the scheduler rules are stable.

Done when:

- Workflow authors can understand what happens after any task failure.
- Failure behavior is covered by docs and tests.
- Existing workflows continue to run without changes.

## Milestone 2: Output And Automation

Goal: make Talos easier to use in long-running local workflows and scripts.

Planned work:

- Add compact and verbose output modes.
- Improve task timing and summary details.
- Add machine-readable output, likely JSON, for automation and CI.
- Add optional Mermaid output to a file.
- Keep human output clean by default.

Suggested implementation order:

1. Add `talos visualize --output <path>`.
2. Define a stable JSON schema for run summaries.
3. Add `talos run --output json`.
4. Add compact and verbose human output modes.
5. Update examples and command docs.

Done when:

- Developers can quickly scan an interactive run.
- Scripts can consume Talos output without parsing human text.
- Documentation shows both human and machine-readable usage.

## Milestone 3: Distribution And Adoption

Goal: make the project easy to install, evaluate, and contribute to.

Planned work:

- Add Homebrew installation support.
- Add contribution guidelines.
- Add issue templates for bugs and feature requests.
- Add a small demo project that uses Talos in a realistic workflow.
- Add a short comparison section explaining when to use Talos instead of `make`, npm scripts, or CI-only pipelines.

Suggested implementation order:

1. Add `CONTRIBUTING.md`.
2. Add issue templates.
3. Add a demo project under `examples/demo`.
4. Add comparison documentation.
5. Add Homebrew distribution once release behavior feels stable.

Done when:

- A new user can install Talos, run a demo, and understand the project in under 10 minutes.
- A contributor can find the test command, coding expectations, and release process without reading source code.

## Milestone 4: `v1.0.0`

Goal: stabilize the workflow schema and CLI behavior.

Planned work:

- Freeze the supported workflow schema for `v1.x`.
- Document compatibility expectations for workflow files.
- Audit defaults, failure behavior, and CLI output for consistency.
- Test the release candidate against real Go, Node.js, Python, Docker, and monorepo workflows.
- Review all examples and docs against the final `v1.0.0` behavior.

Done when:

- The project can support `v1.x` workflows without breaking changes.
- The README, docs, examples, and CLI help all describe the same behavior.
- Release artifacts and checksums are verified from a clean tag.

## Later Ideas

These ideas are useful, but they should wait until the core runner is stable:

- File watching.
- Task caching.
- Conditional task execution.
- Plugin support.
- Web UI.
- CI integrations.
- Remote execution.

Talos should remain small, understandable, and local-first. Features that require servers, databases, or distributed infrastructure should be added only if they preserve that core idea.
