#!/usr/bin/env bash
# Reference sheets (priority 9) are skipped: the API cannot use them as image references.
# Keeps generating missing art in quota-sized batches whenever the minigen account is healthy.
# Usage (repo root): bash tools/art_watch.sh     — stops when nothing is missing.
set -u
cd "$(dirname "$0")/.." || exit 1
. docs/private/secret
BATCH=${BATCH:-30}          # images per batch (minigen allows ~60-100 per 5 hours)
PAUSE=${PAUSE:-9000}        # seconds between batches (2.5 h)
CHECK=${CHECK:-120}         # seconds between account checks while it is down
(cd server && go build -o bin/artgen ./cmd/artgen) || exit 1
missing() { (cd server && ./bin/artgen plan -p 5) | grep -c '^\[ \]'; }
healthy() {
  curl -s -H "x-api-key: $MINIGEN_SERVICE_API_KEY" "$MINIGEN_URL/api/accounts" |
    python3 -c "import json,sys; a=json.load(sys.stdin)['accounts']; sys.exit(0 if any(x['enabled'] and x['status']=='ok' and not x.get('quota_lock') for x in a) else 1)"
}
while [ "$(missing)" -gt 0 ]; do
  if ! healthy; then
    echo "$(date '+%F %T') account not ready; $(missing) images missing"
    sleep "$CHECK"
    continue
  fi
  echo "$(date '+%F %T') batch start ($(missing) missing)"
  (cd server && ./bin/artgen run -c 2 -p 5 -n "$BATCH" && ./bin/artgen prep && ./bin/artgen status)
  echo "$(date '+%F %T') batch done ($(missing) missing)"
  [ "$(missing)" -gt 0 ] && sleep "$PAUSE"
done
echo "$(date '+%F %T') all art generated"
