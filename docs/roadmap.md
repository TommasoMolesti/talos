# Roadmap

Talos is a small, local-first workflow runner for developers who want predictable, dependency-aware task execution without external infrastructure.

The roadmap is intentionally narrow. Talos should become easier to understand before a run, during a run, and after a failure. It should prove that a small Go CLI can have serious execution semantics, clear output, useful portability, and polished documentation without becoming a CI platform.

Patch releases should be reserved for bug fixes, documentation corrections, and small compatibility updates.

## Toward `v1.0.0`: Stable Release

Goal: stabilize the workflow schema and CLI behavior.

Current status:

- The stable workflow schema is documented in [Workflow Configuration](workflows.md).
- Unsupported workflow fields and duplicate workflow keys are rejected with source locations.
- Core execution behavior is covered by tests, including dependency validation, deterministic constrained scheduling, target runs, quiet and verbose output, JSON summaries, retries, timeouts, cancellation, and duplicate dependency rejection.
- Execution semantics docs have been checked against the implementation and tests.
- Bundled examples validate and dry-run successfully.
- Command help, README command summaries, command docs, workflow docs, examples, and release docs describe the same CLI surface.
- The release workflow runs formatting, vet, tests, and bundled example checks before building artifacts.
- Release documentation uses version placeholders until a real release candidate is ready.

Pre-RC work:

- Run and record the [v1 validation plan](v1-validation.md) against real Go, Node.js, Python, Docker, and monorepo projects.

Release-candidate work:

- Build release binaries from a clean tree with `v1.0.0` metadata and verify `talos version`.
- Run the full release-candidate checklist below.
- Review the generated release artifacts before tagging.

Tag-time work:

- Create the `v1.0.0` tag only after the release candidate passes.
- Confirm generated release checksums match the published artifacts.
- Verify `go install github.com/TommasoMolesti/talos@latest` resolves to `v1.0.0`.

Done when:

- The project can support `v1.x` workflows without breaking changes.
- The README, docs, examples, and CLI help all describe the same behavior.
- Release artifacts and checksums are verified from a clean tag.

Release-candidate checklist:

- `gofmt -l .`
- `go vet ./...`
- `go test ./...`
- `talos validate --file examples/go.yaml`
- `talos validate --file examples/node.yaml`
- `talos validate --file examples/python.yaml`
- `talos validate --file examples/docker.yaml`
- `talos validate --file examples/monorepo.yaml`
- `talos validate --file examples/shell.yaml`
- `talos run --file examples/go.yaml --dry-run`
- `talos run --file examples/node.yaml --dry-run`
- `talos run --file examples/python.yaml --dry-run`
- `talos run --file examples/docker.yaml --dry-run`
- `talos run --file examples/monorepo.yaml --dry-run`
- `talos run --file examples/shell.yaml --dry-run`
- `talos visualize --file examples/monorepo.yaml`

## Project Finish Line

Talos should reach a clear `v1.0.0` and then enter maintenance mode. The project is not meant to grow for years through unrelated feature areas.

Remaining finish-line checks:

- Execution semantics are documented and covered by tests.
- README, examples, command docs, and internals docs agree with the implementation.
- Release artifacts can be built and verified from a clean tag.

After `v1.0.0`, Talos should prefer bug fixes, documentation improvements, compatibility updates, and small refinements over new feature areas.
