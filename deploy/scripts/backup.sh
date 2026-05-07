#!/bin/bash
# TenHub バックアップスクリプト
# S3 / Rclone / ローカルにバックアップ

set -euo pipefail
# set -e:
#   途中で失敗したら終了します。
# set -u:
#   未定義の変数を使ったら終了します。
#   例: MYSQL_ROOT_PASSWORD が読み込めていない場合に気づけます。
# set -o pipefail:
#   `mysqldump | gzip` のようなpipeで、左側のmysqldump失敗も検知できます。

# 設定
# BACKUP_ROOT:
#   VPS上でバックアップファイルを保存する親ディレクトリです。
BACKUP_ROOT="/opt/tenhub/backups"

# DATE:
#   バックアップごとに重複しないディレクトリ名・ファイル名を作るための日時です。
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="$BACKUP_ROOT/$DATE"

# RETENTION_DAYS:
#   何日より古いバックアップを削除するかです。
RETENTION_DAYS=30
COMPRESSION_LEVEL=6

# COMPOSE_FILE / ENV_FILE:
#   本番Docker Composeと環境変数ファイルの場所です。
#   ENV_FILEを読み込むことで MYSQL_ROOT_PASSWORD などを使えるようにします。
COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-/opt/tenhub/docker-compose.prod.yml}"
ENV_FILE="${DOCKER_ENV_FILE:-/opt/tenhub/.env.production}"

# UPLOAD_VOLUME_NAME:
#   gatewayのアップロードファイルを保存するDocker volume名です。
#   compose project名が変わるとvolume名も変わるため、必要なら実行時に上書きします。
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

# .env.production をshell変数として読み込みます。
# set -a:
#   読み込んだ変数を子プロセスにも渡せるようにexportします。
# . "$ENV_FILE":
#   envファイルを現在のshellに読み込みます。
# 注意:
#   .env.production は `KEY=value` 形式で書きます。
#   valueに空白を含める場合はquoteが必要です。
if [ -f "$ENV_FILE" ]; then
    set -a
    . "$ENV_FILE"
    set +a
else
    log_error "env file is missing: $ENV_FILE"
    exit 1
fi

# 通知送信
send_notification() {
    # local:
    #   関数の中だけで使う変数です。
    #   外側の変数名とぶつかりにくくします。
    local status=$1
    local message=$2

    # [ -n "$SLACK_WEBHOOK_URL" ]:
    #   文字列が空でなければSlack通知します。
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
# $(...):
#   コマンド結果を変数へ入れます。
# docker compose ps -q db:
#   dbサービスのコンテナIDだけを取得します。
DB_CONTAINER=$(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps -q db 2>/dev/null | head -1)
REDIS_CONTAINER=$(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" ps -q cache 2>/dev/null | head -1)

if [ -z "$DB_CONTAINER" ]; then
    log_error "データベースコンテナが見つかりません"
    send_notification "failed" "データベースコンテナが見つかりません"
    exit 1
fi

# 1. MySQLバックアップ
log_info "MySQLバックアップ中..."
# mysqldump:
#   MySQL全体をSQLとして出力します。
# --single-transaction:
#   InnoDBでロックを抑えて一貫性のあるdumpを取ります。
# --quick:
#   大きいテーブルでも一気にメモリへ載せず、行を順に読みます。
# gzip:
#   SQLを圧縮してディスク使用量を減らします。
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
    # BGSAVE:
    #   Redisにバックグラウンドでdump.rdbを書かせます。
    # docker cp:
    #   コンテナ内のdump.rdbをVPS側のバックアップディレクトリへコピーします。
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
    # Docker volumeはホストの通常パスから直接扱いづらいため、
    # 一時的なalpineコンテナにvolumeをreadonly mountしてtarにします。
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
# tar czf:
#   c=create, z=gzip, f=file指定です。
#   日時ディレクトリを1つのtar.gzにまとめます。
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
# find ... -mtime +30 -delete:
#   30日より古いバックアップファイルを削除します。
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
