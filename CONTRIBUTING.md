# Contributing

Thanks for helping improve Talos. The project is intentionally small, so the best contributions are focused, easy to review, and aligned with the roadmap.

## Local Setup

Talos is a Go CLI. You need Go installed, then you can run all checks from the repository root:

```bash
gofmt -l .
go vet ./...
go test ./...
```

Build the binary locally with:

```bash
go build -o talos .
```

Try a workflow before changing CLI behavior:

```bash
go run . validate --file examples/go.yaml
go run . run --file examples/go.yaml --dry-run
```

## Coding Expectations

- Keep changes PR-sized and scoped to one behavior or documentation improvement.
- Prefer the existing standard-library style and explicit error handling.
- Add or update tests when CLI behavior, validation, scheduling, summaries, or workflow parsing changes.
- Keep examples realistic but runnable as templates; avoid turning the repository into a sample app.
- Update documentation when flags, workflow fields, release behavior, or examples change.

## Pull Requests

Before opening a pull request:

1. Run `gofmt -l .`, `go vet ./...`, and `go test ./...`.
2. Validate any example workflow you changed with `go run . validate --file <path>`.
3. Check `docs/roadmap.md` and explain how the change fits the current release goal.
4. Keep unrelated formatting, generated files, and local binaries out of the diff.

## Release Process

Release instructions live in [docs/releases.md](docs/releases.md). Contributors usually do not need to create tags, but release-facing changes should keep that document, the README, and the CLI version output in agreement.
