#!/usr/bin/env bash
# Build locally, rsync to the staging host, rebuild the container, install the Caddy block.
# Usage: deploy/staging.sh   (needs DEPLOY_KEY and DEPLOY_HOST, e.g. from docs/private/deploy)
set -euo pipefail
cd "$(dirname "$0")/.."
: "${DEPLOY_KEY:?set DEPLOY_KEY}" "${DEPLOY_HOST:?set DEPLOY_HOST}"
OUT=deploy/.build
SSH=(ssh -i "$DEPLOY_KEY" "$DEPLOY_HOST")

rm -rf "$OUT" && mkdir -p "$OUT/bin" "$OUT/client"
(cd server && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -o "../$OUT/bin/server" ./cmd/server)
(cd client && npm run build)
cp -r config maps "$OUT/"
cp -r client/dist "$OUT/client/dist"
rsync -a --exclude='_raw' assets/ "$OUT/assets/"
cp deploy/Dockerfile deploy/docker-compose.yml "$OUT/"

rsync -az --delete -e "ssh -i $DEPLOY_KEY" "$OUT/" "$DEPLOY_HOST:kiemxoaytutien/"
"${SSH[@]}" 'cd ~/kiemxoaytutien && docker compose up -d --build'

# Caddy block: back up, install, validate the full config, reload — or roll back.
scp -i "$DEPLOY_KEY" deploy/kientiem-staging.caddy "$DEPLOY_HOST:/tmp/kientiem-staging.caddy"
"${SSH[@]}" 'set -e; D=/opt/astery/caddy/conf.d; F=$D/kientiem-staging.caddy
  [ -f $F ] && cp $F /tmp/kientiem-staging.caddy.prev
  cp /tmp/kientiem-staging.caddy $F
  if docker exec docker-caddy-1 caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/tmp/kientiem-caddy.log 2>&1; then
    docker exec docker-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
  else
    if [ -f /tmp/kientiem-staging.caddy.prev ]; then cp /tmp/kientiem-staging.caddy.prev $F; else rm -f $F; fi
    echo "caddy validate failed, rolled back (see /tmp/kientiem-caddy.log)"; exit 1
  fi'
echo "deployed: https://kientiem.staging.maveragroups.com"
