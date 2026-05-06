#!/bin/bash
# TenHub ヘルスチェックスクリプト
# 各サービスの稼働状態を監視し、異常時に通知

set -e

# 設定
BASE_URL=${BASE_URL:-"http://localhost"}
TIMEOUT=10
RETRIES=3

# 通知先（環境変数で設定）
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-}
DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL:-}

# 色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ログ関数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 通知送信
send_alert() {
    local service=$1
    local status=$2
    local details=$3

    local message="❌ *Service Alert*\n*Service:* $service\n*Status:* $status\n*Details:* $details\n*Time:* $(date '+%Y-%m-%d %H:%M:%S')"

    # Slack
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{
                \"text\": \"$service: $status\",
                \"blocks\": [{
                    \"type\": \"section\",
                    \"text\": {
                        \"type\": \"mrkdwn\",
                        \"text\": \"$message\"
                    }
                }]
            }" > /dev/null
    fi

    # Discord
    if [ -n "$DISCORD_WEBHOOK_URL" ]; then
        curl -s -X POST "$DISCORD_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"content\": \"$message\"}" > /dev/null
    fi

    # ログファイル
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $service: $status - $details" >> /var/log/tenhub-health.log
}

# HTTPチェック
check_http() {
    local name=$1
    local url=$2
    local expected_code=${3:-200}

    log_info "Checking $name: $url"

    local response=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT "$url" 2>/dev/null || echo "000")

    if [ "$response" = "$expected_code" ] || \
       [ "$response" = "301" ] || [ "$response" = "302" ]; then
        log_info "✓ $name is healthy (HTTP $response)"
        return 0
    else
        log_error "✗ $name is unhealthy (HTTP $response, expected $expected_code)"
        send_alert "$name" "HTTP $response" "URL: $url"
        return 1
    fi
}

# Dockerコンテナチェック
check_container() {
    local name=$1
    local container_name=$2

    if docker ps --format '{{.Names}}' | grep -q "$container_name"; then
        log_info "✓ Container $name is running"
        return 0
    else
        log_error "✗ Container $name is not running"
        send_alert "$name" "Container Down" "Container: $container_name"
        return 1
    fi
}

# ディスク使用率チェック
check_disk() {
    local threshold=${1:-80}

    local usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

    if [ "$usage" -lt "$threshold" ]; then
        log_info "✓ Disk usage: ${usage}%"
        return 0
    else
        log_warn "⚠ Disk usage high: ${usage}%"
        send_alert "Disk" "High Usage" "Usage: ${usage}%"
        return 1
    fi
}

# メモリ使用率チェック
check_memory() {
    local threshold=${1:-90}

    local usage=$(free | awk '/Mem/{printf "%.0f", ($3/$2) * 100.0}')

    if [ "$usage" -lt "$threshold" ]; then
        log_info "✓ Memory usage: ${usage}%"
        return 0
    else
        log_warn "⚠ Memory usage high: ${usage}%"
        send_alert "Memory" "High Usage" "Usage: ${usage}%"
        return 1
    fi
}

# メイン
main() {
    log_info "========================================"
    log_info "TenHub Health Check Starting..."
    log_info "========================================"

    local failed=0

    # 1. コンテナチェック
    log_info "--- Container Checks ---"
    check_container "Nginx" "tenhub-nginx-1" || failed=$((failed + 1))
    check_container "Frontend" "tenhub-frontend-1" || failed=$((failed + 1))
    check_container "Gateway" "tenhub-gateway-1" || failed=$((failed + 1))
    check_container "Auth" "tenhub-auth-1" || failed=$((failed + 1))
    check_container "Wiki" "tenhub-wiki-1" || failed=$((failed + 1))
    check_container "Profile" "tenhub-profile-1" || failed=$((failed + 1))
    check_container "AI" "tenhub-ai-1" || failed=$((failed + 1))
    check_container "MySQL" "tenhub-db-1" || failed=$((failed + 1))
    check_container "Redis" "tenhub-cache-1" || failed=$((failed + 1))

    # 2. HTTPチェック
    log_info "--- HTTP Checks ---"
    check_http "Frontend" "$BASE_URL/" 200 || failed=$((failed + 1))
    check_http "Gateway Health" "$BASE_URL/api/health" 200 || failed=$((failed + 1))
    check_http "Wiki API" "$BASE_URL/api/articles" 200 || failed=$((failed + 1))

    # 3. システムリソースチェック
    log_info "--- System Checks ---"
    check_disk 80 || failed=$((failed + 1))
    check_memory 90 || failed=$((failed + 1))

    # 4. まとめ
    log_info "========================================"
    if [ $failed -eq 0 ]; then
        log_info "✓ All checks passed!"
        log_info "========================================"
        exit 0
    else
        log_error "✗ $failed check(s) failed!"
        log_info "========================================"
        exit 1
    fi
}

# 引数で特定チェックも可能
case "${1:-all}" in
    http)
        check_http "Frontend" "$BASE_URL/" 200
        check_http "Gateway Health" "$BASE_URL/api/health" 200
        ;;
    containers)
        check_container "Nginx" "tenhub-nginx-1"
        check_container "Frontend" "tenhub-frontend-1"
        check_container "Gateway" "tenhub-gateway-1"
        ;;
    disk)
        check_disk 80
        ;;
    memory)
        check_memory 90
        ;;
    *)
        main
        ;;
esac
