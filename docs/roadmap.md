# Roadmap

Talos is a small, local-first workflow runner for developers who want predictable, dependency-aware task execution without external infrastructure.

The roadmap is intentionally narrow. Talos should become easier to understand before a run, during a run, and after a failure. It should prove that a small Go CLI can have serious execution semantics, clear output, useful portability, and polished documentation without becoming a CI platform.

Patch releases should be reserved for bug fixes, documentation corrections, and small compatibility updates.

## `v0.5.0`: Adoption Polish

Goal: make the project easy to install, evaluate, and contribute to.

Work in this release should stay PR-sized and easy to review.

Planned work:

- [x] Add contribution guidelines.
- [x] Add issue templates for bugs and feature requests.
- [x] Improve one existing example so it feels realistic without becoming a demo app.
- [x] Add a short comparison section explaining when to use Talos instead of `make`, npm scripts, or CI-only pipelines.
- [x] Keep install options focused on release binaries, the install script, and `go install`.

Done when:

- A new user can install Talos, run a realistic example, and understand the project in under 10 minutes.
- A contributor can find the test command, coding expectations, and release process without reading source code.
- The project looks finished without depending on package-manager sprawl.

## `v1.0.0`: Stable Release

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
