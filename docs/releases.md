# Release Process

Talos publishes repeatable GitHub releases from version tags. A tag like `v0.1.0` triggers the release workflow, builds Linux, macOS, and Windows binaries, generates `checksums.txt`, and attaches all assets to the GitHub Release.

## Before Tagging

Run the local checks:

```bash
gofmt -l .
go vet ./...
go test ./...
```

Confirm the version output works with release metadata:

```bash
go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o talos .
./talos version
```

## Create The Release

Create and push a version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow publishes:

- Linux binaries for `amd64` and `arm64`
- macOS binaries for `amd64` and `arm64`
- Windows binary for `amd64`
- `checksums.txt`

## After Publishing

Check the GitHub Release page and confirm that all expected assets are attached. Then verify that `go install github.com/TommasoMolesti/talos@latest` resolves to the new tag.
