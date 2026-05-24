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

BINARY="${GOBIN}/talos"

echo "Talos installed to ${BINARY}"

if [ -x "${BINARY}" ]; then
	"${BINARY}" version
fi

if [ "$(command -v talos 2>/dev/null || true)" = "${BINARY}" ]; then
	exit 0
fi

echo "Talos is installed, but ${BINARY} is not the talos found first in your PATH."
echo "Add it for this shell with:"
echo "  export PATH=\"${GOBIN}:\$PATH\""
echo
echo "For zsh, persist it with:"
echo "  echo 'export PATH=\"${GOBIN}:\$PATH\"' >> ~/.zshrc"
echo "  source ~/.zshrc"
