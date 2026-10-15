#!/usr/bin/env bash
# Builds the web page of the desk into a directory (default: site/dist):
# the page, its pictures, drawn by the desk, and the desk itself, compiled
# to WebAssembly. SITE is the address the page is served at, for the
# pictures a link to it shows.
#
#	./scripts/site.sh
#	SITE=https://example.com ./scripts/site.sh out
#	python3 -m http.server -d site/dist   # then open http://localhost:8000
set -euo pipefail
cd "$(dirname "$0")/.."

out=${1:-site/dist}
site=${SITE:-http://localhost:8000}
rm -rf "$out"
mkdir -p "$out"
cp -R site/index.html site/site.css site/site.js site/llms.txt site/_headers site/fonts "$out/"

go run ./examples/desktop -site "$out"

# The desk, built as mygo build builds for browsers, without the
# inspector; its file and Go's loader named by their contents, to be kept
# for good by caches.
GOOS=js GOARCH=wasm go build -trimpath -tags mygo_noinspector \
  -ldflags "-s -w -X github.com/egoist/mygo.production=1" -o "$out/desk.wasm" ./examples/desktop
hash() { shasum -a 256 "$1" | cut -c1-12; }
wasm=desk.$(hash "$out/desk.wasm").wasm
mv "$out/desk.wasm" "$out/$wasm"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$out/wasm_exec.js"
exec_js=wasm_exec.$(hash "$out/wasm_exec.js").js
mv "$out/wasm_exec.js" "$out/$exec_js"
size=$(wc -c < "$out/$wasm" | tr -d ' ')

sed -i.bak -e "s|__SITE__|$site|g" -e "s|__WASM__|$wasm|g" -e "s|__WASM_SIZE__|$size|g" \
  -e "s|__WASM_EXEC__|$exec_js|g" "$out/index.html"
rm "$out/index.html.bak"
echo "$out: $wasm ($size bytes)"
