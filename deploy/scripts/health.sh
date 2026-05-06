#!/bin/bash
set -e

# Simple production health check.
# Run on VPS:
#   /opt/tenhub/scripts/health.sh

BASE_URL="${BASE_URL:-http://localhost}"
COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-/opt/tenhub/docker-compose.prod.yml}"
DISK_THRESHOLD="${HEALTH_CHECK_DISK_THRESHOLD:-80}"
MEMORY_THRESHOLD="${HEALTH_CHECK_MEMORY_THRESHOLD:-90}"

ok() { echo "[OK] $1"; }
fail() { echo "[NG] $1"; exit 1; }

check_service() {
    service="$1"
    id="$(docker compose -f "$COMPOSE_FILE" ps -q "$service" 2>/dev/null || true)"
    [ -n "$id" ] || fail "$service container is missing"
    running="$(docker inspect -f '{{.State.Running}}' "$id" 2>/dev/null || echo false)"
    [ "$running" = "true" ] || fail "$service container is not running"
    ok "$service is running"
}

check_http() {
    code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$BASE_URL/" || echo 000)"
    case "$code" in
        200|301|302|304) ok "HTTP check passed: $code" ;;
        *) fail "HTTP check failed: $code" ;;
    esac
}

check_disk() {
    usage="$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')"
    [ "$usage" -lt "$DISK_THRESHOLD" ] || fail "disk usage is high: ${usage}%"
    ok "disk usage: ${usage}%"
}

check_memory() {
    usage="$(free | awk '/Mem/{printf "%.0f", ($3/$2) * 100.0}')"
    [ "$usage" -lt "$MEMORY_THRESHOLD" ] || fail "memory usage is high: ${usage}%"
    ok "memory usage: ${usage}%"
}

for service in nginx frontend gateway auth wiki profile ai db cache; do
    check_service "$service"
done

check_http
check_disk
check_memory

ok "all checks passed"
