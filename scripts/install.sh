#!/usr/bin/env sh
set -eu

VERSION="${VER:-${1:-latest}}"
MODULE="github.com/TommasoMolesti/talos"

echo "Installing Talos ${VERSION}..."
go install "${MODULE}@${VERSION}"

GOBIN="$(go env GOBIN)"
if [ -z "${GOBIN}" ]; then
	GOBIN="$(go env GOPATH)/bin"
fi

echo "Talos installed to ${GOBIN}/talos"

if command -v talos >/dev/null 2>&1; then
	talos version
	exit 0
fi

echo "Talos is installed, but ${GOBIN} is not in your PATH."
echo "Add it for this shell with:"
echo "  export PATH=\"\$PATH:${GOBIN}\""
echo
echo "For zsh, persist it with:"
echo "  echo 'export PATH=\"\$PATH:${GOBIN}\"' >> ~/.zshrc"
echo "  source ~/.zshrc"
