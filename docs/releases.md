# Release Process

Talos publishes repeatable GitHub releases from version tags. A tag like `vX.Y.Z` triggers the release workflow, builds Linux, macOS, and Windows binaries, generates `checksums.txt`, and attaches all assets to the GitHub Release. Release tags must match `vX.Y.Z`.

## Before Tagging

Run the local checks:

```bash
gofmt -l .
go vet ./...
go test ./...
```

The release workflow runs these checks again before building release artifacts.

Confirm the version output works with release metadata:

```bash
VERSION=vX.Y.Z
go build -ldflags "-X main.version=${VERSION} -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o talos .
./talos version
```

Validate the bundled example workflows:

```bash
for file in examples/*.yaml; do
  ./talos validate --file "$file"
  ./talos run --file "$file" --dry-run
done
./talos visualize --file examples/monorepo.yaml
```

The release workflow also validates and dry-runs the bundled examples, then checks monorepo visualization, before building artifacts.

Before creating a `v1.0.0` release candidate, complete the [v1 validation plan](v1-validation.md) against real projects. Every validation row must be `pass` or have a documented, intentional exclusion; do not tag with `pending`, `blocked`, or failing rows.

## Create The Release

Create and push a version tag:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

The release workflow publishes:

- Linux binaries for `amd64` and `arm64`
- macOS binaries for `amd64` and `arm64`
- Windows binary for `amd64`
- `checksums.txt`

## After Publishing

Check the GitHub Release page and confirm that all expected assets are attached.

Download `checksums.txt` and the expected binary for your platform, then verify the checksum locally:

```bash
sha256sum --check checksums.txt
```

On macOS, use:

```bash
shasum -a 256 -c checksums.txt
```

Run the downloaded binary and confirm it reports the release tag:

```bash
./talos version
```

Then verify that `go install github.com/TommasoMolesti/talos@latest` resolves to the new tag.
