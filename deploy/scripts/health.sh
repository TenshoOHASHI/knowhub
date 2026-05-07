#!/bin/bash
# bashで実行するヘルスチェックスクリプトです。
# `/bin/sh` ではなく `/bin/bash` を指定しているのは、
# 関数定義や `[[ ... ]]` 等へ拡張しやすくするためです。

set -e
# set -e:
#   途中のコマンドが失敗したら、その時点でスクリプトを終了します。
#   本番確認では「失敗しているのに最後まで進む」方が危険なので有効にします。

# Simple production health check.
# Run on VPS:
#   /opt/tenhub/scripts/health.sh

# BASE_URL:
#   HTTP確認でアクセスするURLです。
#   何も指定しなければVPS内から nginx の80番へ `http://localhost/` で確認します。
#   ドメインで確認したい場合は次のように実行します。
#     BASE_URL=https://example.com /opt/tenhub/scripts/health.sh
BASE_URL="${BASE_URL:-http://localhost}"

# COMPOSE_FILE:
#   本番Docker Composeファイルの場所です。
#   DOCKER_COMPOSE_FILEを外から渡すと、別のcomposeファイルでも確認できます。
COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-/opt/tenhub/docker-compose.prod.yml}"

# ENV_FILE:
#   docker-compose.prod.yml 内の ${...} 変数を解決するためのenvファイルです。
#   image名やDB設定などを読むため、本番では .env.production を明示します。
ENV_FILE="${DOCKER_ENV_FILE:-/opt/tenhub/.env.production}"

# 閾値:
#   diskが80%以上、memoryが90%以上なら異常として終了します。
#   必要なら実行時に上書きできます。
#     HEALTH_CHECK_DISK_THRESHOLD=90 /opt/tenhub/scripts/health.sh
DISK_THRESHOLD="${HEALTH_CHECK_DISK_THRESHOLD:-80}"
MEMORY_THRESHOLD="${HEALTH_CHECK_MEMORY_THRESHOLD:-90}"

# ok/fail は小さな共通関数です。
# `$1` は関数に渡された1つ目の引数です。
ok() { echo "[OK] $1"; }
fail() {
    echo "[NG] $1"
    exit 1
}

check_service() {
    # `$1` は check_service nginx のように渡されたサービス名です。
    service="$1"

    # command substitution:
    #   `$(...)` の中のコマンド結果を文字列として変数に入れます。
    # docker compose ps -q:
    #   指定サービスのコンテナIDだけを出します。
    # 2>/dev/null:
    #   エラーメッセージを画面に出さず捨てます。
    # || true:
    #   コンテナが無くても set -e で即終了しないようにします。
    id="$(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps -q "$service" 2>/dev/null || true)"

    # [ -n "$id" ]:
    #   idが空でなければtrueです。
    #   空ならコンテナが存在しないのでfailします。
    [ -n "$id" ] || fail "$service container is missing"

    # docker inspect:
    #   コンテナの詳細情報から Running=true/false だけを取り出します。
    running="$(docker inspect -f '{{.State.Running}}' "$id" 2>/dev/null || echo false)"

    # [ "$running" = "true" ]:
    #   文字列比較です。true以外なら起動していない扱いにします。
    [ "$running" = "true" ] || fail "$service container is not running"

    ok "$service is running"
}

check_http() {
    # curl:
    #   -s は進捗を非表示
    #   -o /dev/null は本文を捨てる
    #   -w '%{http_code}' はHTTPステータスコードだけ表示
    #   --max-time 10 は10秒でtimeout
    code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$BASE_URL/" || echo 000)"

    # case:
    #   複数条件の分岐です。
    #   200/301/302/304 は正常応答として扱います。
    case "$code" in
        200|301|302|304) ok "HTTP check passed: $code" ;;
        *) fail "HTTP check failed: $code" ;;
    esac
}

check_disk() {
    # df -h /:
    #   ルートディスクの使用量を確認します。
    # awk 'NR==2 {print $5}':
    #   2行目の5列目、つまり使用率 `xx%` を取り出します。
    # sed 's/%//':
    #   数値比較できるように `%` を削除します。
    usage="$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')"

    # -lt:
    #   less than の意味です。左辺が右辺より小さければtrueです。
    [ "$usage" -lt "$DISK_THRESHOLD" ] || fail "disk usage is high: ${usage}%"
    ok "disk usage: ${usage}%"
}

check_memory() {
    # free:
    #   メモリ使用状況を表示します。
    # awk:
    #   Mem行から used / total を計算して、使用率を整数で出します。
    usage="$(free | awk '/Mem/{printf "%.0f", ($3/$2) * 100.0}')"
    [ "$usage" -lt "$MEMORY_THRESHOLD" ] || fail "memory usage is high: ${usage}%"
    ok "memory usage: ${usage}%"
}

# for:
#   対象サービスを1つずつ check_service に渡します。
#   この一覧は docker-compose.prod.yml の services と揃えます。
for service in nginx frontend gateway auth wiki profile ai db cache; do
    check_service "$service"
done

# コンテナの起動確認だけでなく、HTTP・ディスク・メモリも確認します。
check_http
check_disk
check_memory

ok "all checks passed"
