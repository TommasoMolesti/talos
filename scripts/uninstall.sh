#!/usr/bin/env sh
set -eu

GOBIN="$(go env GOBIN)"
if [ -z "${GOBIN}" ]; then
	GOBIN="$(go env GOPATH)/bin"
fi

BINARY="${GOBIN}/talos"

if [ ! -f "${BINARY}" ]; then
	echo "Talos is not installed at ${BINARY}"
	exit 0
fi

rm "${BINARY}"
echo "Removed ${BINARY}"
echo "Talos does not create background services or global config files."
