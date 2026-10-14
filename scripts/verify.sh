#!/bin/sh
# Runs every check of the repository: go vet, gofmt, the build, and every
# test — snapshots compared pixel by pixel with their golden images, checks
# against docs/spec.md, interaction tests and the desktop example's smoke
# test. The last line is "PASSED x/y". Works from any directory, offline.
cd "$(dirname "$0")/.." || exit 2
total=$(grep -cv '^#' scripts/checks.txt 2>/dev/null || echo 1)
tmp=$(mktemp -d 2>/dev/null || mktemp -d -t verify)
trap 'rm -rf "$tmp"' EXIT
if ! go build -o "$tmp/verify" ./internal/verify; then
	echo "the verifier did not build"
	echo "PASSED 0/$total"
	exit 1
fi
"$tmp/verify" "$@"
status=$?
if [ $status -gt 1 ]; then
	# The verifier crashed before reporting: nothing passed.
	echo "PASSED 0/$total"
fi
exit $status
