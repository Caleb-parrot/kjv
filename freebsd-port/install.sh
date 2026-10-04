#!/bin/sh
# Build and install kjv-tui on GhostBSD / FreeBSD.
# Run from the repository root: sh freebsd-port/install.sh

set -eu

cd "$(dirname "$0")/.."

if ! command -v go >/dev/null 2>&1; then
	echo "Go is not installed. On GhostBSD run:  sudo pkg install go git xclip" >&2
	exit 1
fi

echo "Building kjv-tui..."
go build -o kjv-tui .

PREFIX="${PREFIX:-/usr/local}"
echo "Installing to ${PREFIX} (needs root)..."
install -d "${DESTDIR}${PREFIX}/bin"
install -d "${DESTDIR}${PREFIX}/share/applications"
install -m 755 kjv-tui "${DESTDIR}${PREFIX}/bin/kjv-tui"
install -m 644 kjv-tui.desktop "${DESTDIR}${PREFIX}/share/applications/kjv-tui.desktop"

echo "Installed ${PREFIX}/bin/kjv-tui"
echo "Open it from the menu as KJV, or run: kjv-tui"
