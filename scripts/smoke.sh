#!/usr/bin/env bash
# End-to-end smoke test:
#   1. wait for api rollout
#   2. port-forward api:80 -> localhost:${PORT:-18080}
#   3. wait until the forwarded health endpoint responds
#   4. POST jobs
#   5. wait for worker to drain
#   6. assert api_jobs_done >= JOBS in /metrics
set -euo pipefail

NS="${NS:-app}"
RELEASE="${RELEASE:-app}"
JOBS="${JOBS:-5}"
TIMEOUT="${TIMEOUT:-60}"
PORT="${PORT:-18080}"
PF_LOG="${TMPDIR:-/tmp}/devsecops-port-forward-${PORT}.log"
API_URL="http://127.0.0.1:${PORT}"

echo "[smoke] waiting for api rollout"
kubectl -n "$NS" rollout status "deploy/${RELEASE}-api" --timeout=120s

echo "[smoke] port-forwarding"
kubectl -n "$NS" port-forward "svc/${RELEASE}-api" "${PORT}:80" >"$PF_LOG" 2>&1 &
PF_PID=$!

cleanup() {
  kill "$PF_PID" 2>/dev/null || true
  wait "$PF_PID" 2>/dev/null || true
}
trap cleanup EXIT

echo "[smoke] waiting for forwarded health endpoint"
READY=false
for _ in $(seq 1 15); do
  if curl -fsS --max-time 1 "${API_URL}/healthz" >/dev/null; then
    READY=true
    break
  fi

  if ! kill -0 "$PF_PID" 2>/dev/null; then
    echo "[smoke] FAIL: kubectl port-forward exited"
    cat "$PF_LOG"
    exit 1
  fi
  sleep 1
done

if [[ "$READY" != true ]]; then
  echo "[smoke] FAIL: api did not become reachable through port-forward"
  cat "$PF_LOG"
  exit 1
fi

echo "[smoke] posting $JOBS jobs"
for i in $(seq 1 "$JOBS"); do
  curl -sf -X POST -H "Content-Type: application/json" \
    -d "{\"n\":$i}" "${API_URL}/api/v1/jobs" > /dev/null
done

echo "[smoke] waiting up to ${TIMEOUT}s for worker to drain"
for _ in $(seq 1 "$TIMEOUT"); do
  done_n=$(curl -sf "${API_URL}/metrics" | awk '/^api_jobs_done /{print $2}')
  if [[ "${done_n:-0}" -ge "$JOBS" ]]; then
    echo "[smoke] OK: api_jobs_done=$done_n"
    exit 0
  fi
  sleep 1
done

echo "[smoke] FAIL: drain did not complete within ${TIMEOUT}s"
curl -s "${API_URL}/metrics" | grep ^api_
exit 1
