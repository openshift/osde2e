#!/bin/bash
#
# A shellcheck installer/wrapper.
#

SCRIPT_DIR="$(dirname "$0")"
scversion="stable" # or "v0.4.7", or "latest"
[ ! -d "$SCRIPT_DIR/shellcheck-${scversion?}" ] && wget -qO- "https://github.com/koalaman/shellcheck/releases/download/${scversion?}/shellcheck-${scversion?}.linux.x86_64.tar.xz" | tar -xJv -C "$SCRIPT_DIR"
"$SCRIPT_DIR/shellcheck-${scversion?}/shellcheck" "$@"
