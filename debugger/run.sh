#!/usr/bin/env bash
# run.sh launches the frame debugger, hiding the GOWORK=off / module-directory
# dance. The debugger is a standalone module (it pins a newer Go than the
# workspace and carries the repo's only cgo dependency), so it must be run with
# the workspace disabled and from inside its own directory.
#
# Usage:
#   ./debugger/run.sh                              # default ROM (Pilotwings 64)
#   ./debugger/run.sh path/to/game.z64             # a specific ROM
#   ./debugger/run.sh path/to/game.z64 -state s.st # extra flags pass through
#
# A relative ROM path is resolved against the directory you run this from.
set -euo pipefail

orig="$PWD"
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo="$(cd "$here/.." && pwd)"

rom="$repo/games/pilotwings-64-n64/image/Pilotwings 64 (USA).z64"

# A leading non-flag argument overrides the ROM; everything after passes through.
if [[ $# -gt 0 && "$1" != -* ]]; then
	rom="$1"
	shift
fi

# Make a relative ROM path absolute so it survives the cd into the module.
case "$rom" in
	/*) ;;
	*) rom="$orig/$rom" ;;
esac

cd "$here"
exec env GOWORK=off go run . -image "$rom" "$@"
