#!/usr/bin/env bash
# Subsets the fonts of the web page (site/fonts) from a release of Platinum
# Fonts to the characters its page and script use, in two faces — Platinum
# Sans Bold and Platinum Sans Small Regular, the faces of SystemFont and
# AppFont — and four regional forms. Run it again after changing the
# page's words. Needs pyftsubset (pip install fonttools brotli) and gh or
# curl.
set -euo pipefail
cd "$(dirname "$0")/.."

version=v0.2.0
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
zip=platinum-fonts-$version-woff2.zip
curl -fsSL -o "$work/$zip" "https://github.com/firede/platinum-fonts/releases/download/$version/$zip"
unzip -q "$work/$zip" -d "$work"
fonts=$work/platinum-fonts-$version-woff2

# Printable ASCII, and every character of the page and its script.
text=$work/text.txt
{ printf '%s' ' !"#$%&'"'"'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_`abcdefghijklmnopqrstuvwxyz{|}~'; cat site/index.html site/site.js; } > "$text"

mkdir -p site/fonts
for region in SC TC JP KR; do
  lower=$(printf '%s' "$region" | tr '[:upper:]' '[:lower:]')
  for face in "Sans:Bold:sans-bold" "SansSmall:Regular:sans-small"; do
    IFS=: read -r family style name <<< "$face"
    pyftsubset "$fonts/Platinum${family}${region}-${style}.woff2" \
      --text-file="$text" --flavor=woff2 --layout-features='' --no-hinting \
      --output-file="site/fonts/$name-$lower.woff2"
  done
done
install -m 644 "$fonts/OFL.txt" site/fonts/OFL.txt
ls -l site/fonts
