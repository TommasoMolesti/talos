# Roadmap

Talos is a small, local-first workflow runner for developers who want predictable, dependency-aware task execution without external infrastructure.

The roadmap is intentionally narrow. Talos should become easier to understand before a run, during a run, and after a failure. It should prove that a small Go CLI can have serious execution semantics, clear output, useful portability, and polished documentation without becoming a CI platform.

Patch releases should be reserved for bug fixes, documentation corrections, and small compatibility updates.

## `v1.0.0`: Stable Release

Goal: stabilize the workflow schema and CLI behavior.

Planned work:

- Freeze the supported workflow schema for `v1.x`. (In progress: unsupported workflow fields are rejected.)
- Document compatibility expectations for workflow files. (In progress: workflow compatibility notes are documented.)
- Audit defaults, failure behavior, and CLI output for consistency.
- Test the release candidate against real Go, Node.js, Python, Docker, and monorepo workflows.
- Review all examples and docs against the final `v1.0.0` behavior.

Done when:

- The project can support `v1.x` workflows without breaking changes.
- The README, docs, examples, and CLI help all describe the same behavior.
- Release artifacts and checksums are verified from a clean tag.

## Project Finish Line

Talos should reach a clear `v1.0.0` and then enter maintenance mode. The project is not meant to grow for years through unrelated feature areas.

The project is considered complete when:

- The workflow schema is stable and documented.
- Execution semantics are documented and covered by tests.
- Shell configuration is simple and predictable.
- Run reporting makes parallel execution easy to follow.
- README, examples, command docs, and internals docs agree with the implementation.
- Release artifacts can be built and verified from a clean tag.

After `v1.0.0`, Talos should prefer bug fixes, documentation improvements, compatibility updates, and small refinements over new feature areas.
