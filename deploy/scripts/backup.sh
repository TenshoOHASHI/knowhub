#!/bin/bash
# TenHub バックアップスクリプト
# S3 / Rclone / ローカルにバックアップ

set -e

# 設定
BACKUP_ROOT="/opt/tenhub/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="$BACKUP_ROOT/$DATE"
RETENTION_DAYS=30
COMPRESSION_LEVEL=6
COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-/opt/tenhub/docker-compose.prod.yml}"
UPLOAD_VOLUME_NAME="${UPLOAD_VOLUME_NAME:-tenhub-prod_gateway-uploads}"

# S3設定（オプション）
S3_BUCKET=${S3_BUCKET:-}
S3_PATH=${S3_PATH:-s3://tenhub-backups}

# 通知用（オプション）
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-}

# 色
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 通知送信
send_notification() {
    local status=$1
    local message=$2

    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        local emoji="✅"
        [ "$status" = "failed" ] && emoji="❌"

        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{
                \"text\": \"$emoji TenHub Backup $status\",
                \"blocks\": [{
                    \"type\": \"section\",
                    \"text\": {
                        \"type\": \"mrkdwn\",
                        \"text\": \"*$status*$message\n*Date:* $(date '+%Y-%m-%d %H:%M:%S')\"
                    }
                }]
            }" > /dev/null
    fi
}

# バックアップ開始
log_info "バックアップを開始します..."
mkdir -p "$BACKUP_DIR"

# Dockerコンテナ名取得
DB_CONTAINER=$(docker compose -f "$COMPOSE_FILE" ps -q db 2>/dev/null | head -1)
REDIS_CONTAINER=$(docker compose -f "$COMPOSE_FILE" ps -q cache 2>/dev/null | head -1)

if [ -z "$DB_CONTAINER" ]; then
    log_error "データベースコンテナが見つかりません"
    send_notification "failed" "データベースコンテナが見つかりません"
    exit 1
fi

# 1. MySQLバックアップ
log_info "MySQLバックアップ中..."
docker exec "$DB_CONTAINER" mysqldump \
    -u root \
    -p"${MYSQL_ROOT_PASSWORD}" \
    --single-transaction \
    --quick \
    --lock-tables=false \
    --all-databases \
    2>/dev/null | gzip > "$BACKUP_DIR/mysql.sql.gz"

if [ $? -eq 0 ]; then
    SIZE=$(du -h "$BACKUP_DIR/mysql.sql.gz" | cut -f1)
    log_info "MySQLバックアップ完了: $SIZE"
else
    log_error "MySQLバックアップ失敗"
    send_notification "failed" "MySQLバックアップ失敗"
    exit 1
fi

# 2. Redisバックアップ
log_info "Redisバックアップ中..."
if [ -n "$REDIS_CONTAINER" ]; then
    docker exec "$REDIS_CONTAINER" redis-cli BGSAVE
    docker cp "$REDIS_CONTAINER:/data/dump.rdb" "$BACKUP_DIR/redis.rdb"
    gzip "$BACKUP_DIR/redis.rdb"
    log_info "Redisバックアップ完了"
else
    log_info "Redisコンテナが見つかりません、スキップ"
fi

# 3. アップロードファイルバックアップ
log_info "アップロードファイルバックアップ中..."
if docker volume inspect "$UPLOAD_VOLUME_NAME" >/dev/null 2>&1; then
    docker run --rm \
        -v "$UPLOAD_VOLUME_NAME:/data:ro" \
        -v "$BACKUP_DIR":/backup \
        alpine tar czf /backup/uploads.tar.gz -C /data .
    log_info "アップロードファイルバックアップ完了"
else
    log_info "アップロード用volumeが見つかりません、スキップ: $UPLOAD_VOLUME_NAME"
fi

# 4. JWT鍵バックアップ
log_info "JWT鍵バックアップ中..."
mkdir -p "$BACKUP_DIR/keys"
cp -r /opt/tenhub/services/auth/keys/* "$BACKUP_DIR/keys/" 2>/dev/null || true

# 5. 環境変数バックアップ
log_info "環境変数バックアップ中..."
cp /opt/tenhub/.env.production "$BACKUP_DIR/env.production" 2>/dev/null || true

# 6. アーカイブ作成
log_info "バックアップアーカイブ作成中..."
cd "$BACKUP_ROOT"
tar czf "tenhub_backup_$DATE.tar.gz" "$DATE"
rm -rf "$DATE"

# 7. S3アップロード（オプション）
if [ -n "$S3_BUCKET" ] && command -v aws &> /dev/null; then
    log_info "S3にアップロード中..."
    aws s3 cp "tenhub_backup_$DATE.tar.gz" "$S3_PATH/" \
        --storage-class STANDARD_IA \
        --metadata "date=$DATE"

    if [ $? -eq 0 ]; then
        log_info "S3アップロード完了"

        # 古いバックアップをS3から削除
        log_info "古いバックアップを清理中..."
        aws s3 ls "$S3_PATH/" | while read -r line; do
            file_date=$(echo "$line" | awk '{print $4}' | grep -oP '\d{8}_\d{6}' | head -1)
            if [ -n "$file_date" ]; then
                file_time=$(date -d "${file_date:0:8} ${file_date:9:2}:${file_date:11:2}:${file_date:13:2}" +%s 2>/dev/null || echo "0")
                cutoff_time=$(date -d "$RETENTION_DAYS days ago" +%s)
                if [ "$file_time" -lt "$cutoff_time" ]; then
                    filename=$(echo "$line" | awk '{print $4}')
                    aws s3 rm "$S3_PATH/$filename"
                    log_info "古いバックアップ削除: $filename"
                fi
            fi
        done
    else
        log_error "S3アップロード失敗"
    fi
fi

# 8. 古いバックアップ削除
log_info "古いバックアップを清理中..."
find "$BACKUP_ROOT" -name "tenhub_backup_*.tar.gz" -mtime +$RETENTION_DAYS -delete

# 9. まとめ
TOTAL_SIZE=$(du -sh "$BACKUP_ROOT" | cut -f1)
BACKUP_COUNT=$(find "$BACKUP_ROOT" -name "tenhub_backup_*.tar.gz" | wc -l)

log_info "バックアップ完了!"
log_info "バックアップサイズ: $TOTAL_SIZE"
log_info "バックアップ数: $BACKUP_COUNT"
log_info "保持期間: $RETENTION_DAYS 日"

send_notification "success" "バックアップ完了\nサイズ: $TOTAL_SIZE\nファイル数: $BACKUP_COUNT"

# ヘルスチェック
log_info "バックアップファイル検証中..."
if tar tzf "tenhub_backup_$DATE.tar.gz" >/dev/null 2>&1; then
    log_info "バックアップファイル検証OK"
else
    log_error "バックアップファイルが破損しています"
    send_notification "failed" "バックアップファイル破損"
    exit 1
fi
