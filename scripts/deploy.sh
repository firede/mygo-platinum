#!/usr/bin/env bash
# Builds the web page of the desk for its address and deploys it to
# Cloudflare, as wrangler.jsonc says. Needs npx and wrangler, logged in
# (npx wrangler login) to the account of the domain.
#
#	./scripts/deploy.sh
set -euo pipefail
cd "$(dirname "$0")/.."

SITE=https://mygo-platinum.firede.com ./scripts/site.sh site/dist
npx wrangler deploy
