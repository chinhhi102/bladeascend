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
# Chép `assets/dist` (bản phát hành do asset-pipeline dựng, spec §13.7), KHÔNG chép `assets/`.
# `assets/` từ 1/11 chỉ còn `src/` (art gốc) và `dist/`; chép nguyên thì server tìm
# `/app/assets/manifest.json` không thấy và staging chạy KHÔNG MỘT ẢNH nào.
[ -f assets/dist/manifest.json ] || { echo "thiếu assets/dist/manifest.json — npm run build hỏng?"; exit 1; }
rsync -a assets/dist/ "$OUT/assets/"
cp deploy/Dockerfile deploy/docker-compose.yml "$OUT/"

rsync -az --delete -e "ssh -i $DEPLOY_KEY" "$OUT/" "$DEPLOY_HOST:bladeascend/"
"${SSH[@]}" 'cd ~/bladeascend && docker compose up -d --build'

# Caddy block: back up, install, validate the full config, reload — or roll back.
scp -i "$DEPLOY_KEY" deploy/bladeascend-staging.caddy "$DEPLOY_HOST:/tmp/bladeascend-staging.caddy"
"${SSH[@]}" 'set -e; D=/opt/astery/caddy/conf.d; F=$D/bladeascend-staging.caddy
  [ -f $F ] && cp $F /tmp/bladeascend-staging.caddy.prev
  cp /tmp/bladeascend-staging.caddy $F
  if docker exec docker-caddy-1 caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile >/tmp/bladeascend-caddy.log 2>&1; then
    docker exec docker-caddy-1 caddy reload --config /etc/caddy/Caddyfile --adapter caddyfile
  else
    if [ -f /tmp/bladeascend-staging.caddy.prev ]; then cp /tmp/bladeascend-staging.caddy.prev $F; else rm -f $F; fi
    echo "caddy validate failed, rolled back (see /tmp/bladeascend-caddy.log)"; exit 1
  fi'
echo "deployed: https://bladeascend.staging.maveragroups.com"
