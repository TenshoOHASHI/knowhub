#!/bin/bash
echo "This legacy all-in-one setup script is deprecated."
echo "Use doc/lighthouse-setup.md and step-by-step commands instead."
exit 1

# VPS初期設定スクリプト
# Tencent Cloud Lighthouse / Ubuntu向け
# 使い方: curl -fsSL https://raw.githubusercontent.com/.../setup-vps.sh | bash

set -e

# 色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# rootチェック
if [ "$EUID" -ne 0 ]; then
    log_error "このスクリプトはroot権限で実行してください"
    exit 1
fi

log_info "TenHub VPS 初期設定を開始します..."

# 1. システム更新
log_info "システムパッケージを更新しています..."
apt-get update -qq
apt-get upgrade -y -qq
apt-get autoremove -y -qq

# 2. 必要パッケージインストール
log_info "必要なパッケージをインストールしています..."
apt-get install -y -qq \
    curl \
    wget \
    git \
    ufw \
    fail2ban \
    unzip \
    htop \
    ncdu \
    jq \
    certbot \
    python3-certbot-nginx \
    ca-certificates \
    gnupg \
    lsb-release

# 3. SSH設定
log_info "SSH設定を変更しています..."

# SSHポート変更（ランダムに）
SSH_PORT=$(shuf -i 20000-40000 -n 1)
sed -i "s/#Port 22/Port $SSH_PORT/" /etc/ssh/sshd_config
sed -i "s/Port 22/Port $SSH_PORT/" /etc/ssh/sshd_config

# rootログイン禁止
sed -i 's/#PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config
sed -i 's/PermitRootLogin yes/PermitRootLogin no/' /etc/ssh/sshd_config

# パスワード認証禁止（鍵認証のみ）
sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

log_info "SSHポートを $SSH_PORT に変更しました（後でファイアウォール設定が必要）"
log_warn "SSH鍵認証を設定してください。設定後、パスワード認証が無効化されます。"
log_warn "現在のSSHポート: $SSH_PORT"
log_warn "このセッションを閉じる前に、新しいポートで接続できることを確認してください！"

# 4. ファイアウォール設定
log_info "ファイアウォールを設定しています..."

# デフォルトで拒否
ufw --force enable
ufw default deny incoming
ufw default allow outgoing

# 必要なポートを許可
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'
ufw allow $SSH_PORT/tcp comment "SSH (custom port)"

# 設定適用
ufw --force reload
ufw status numbered

log_info "ファイアウォール設定完了（SSHポート $SSH_PORT を許可済み）"

# 5. Fail2Ban設定
log_info "Fail2Banを設定しています..."
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5
destemail = root@localhost
sendername = Fail2Ban

[sshd]
enabled = true
port = ssh
logpath = /var/log/auth.log
maxretry = 3
EOF

systemctl enable fail2ban
systemctl restart fail2ban
log_info "Fail2Ban設定完了"

# 6. Dockerインストール
log_info "Dockerをインストールしています..."
if ! command -v docker &> /dev/null; then
    mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    log_info "Dockerインストール完了"
else
    log_info "Dockerは既にインストールされています"
fi

# 7. Dockerユーザー設定
DEPLOY_USER=${SUDO_USER:-lighthouse}
if id "$DEPLOY_USER" &>/dev/null; then
    usermod -aG docker $DEPLOY_USER
    log_info "ユーザー $DEPLOY_USER をdockerグループに追加しました"
fi

# 8. スワップ設定（メモリ節約）
log_info "スワップを設定しています..."
if ! swapon --show | grep -q '/swapfile'; then
    fallocate -l 2G /swapfile
    chmod 600 /swapfile
    mkswap /swapfile
    swapon /swapfile
    echo '/swapfile none swap sw 0 0' >> /etc/fstab
    sysctl vm.swappiness=10
    echo 'vm.swappiness=10' >> /etc/sysctl.conf
    log_info "2GBスワップ作成完了"
else
    log_info "スワップは既に設定されています"
fi

# 9. デプロイディレクトリ作成
DEPLOY_PATH="/opt/tenhub"
mkdir -p $DEPLOY_PATH
mkdir -p $DEPLOY_PATH/backups
mkdir -p $DEPLOY_PATH/logs

# 10. .env.productionテンプレート作成
log_info ".env.production テンプレートを作成しています..."
cat > $DEPLOY_PATH/.env.production.example << 'EOF'
# サーバー設定
PUBLIC_ORIGIN=https://your-domain.com
SERVER_NAME=your-domain.com

# データベース
MYSQL_ROOT_PASSWORD=change-me-root-$(openssl rand -hex 16)
MYSQL_DATABASE=knowhub
MYSQL_USER=knowhub
MYSQL_PASSWORD=change-me-app-$(openssl rand -hex 16)

# アプリケーション
LOG_LEVEL=info

# CORS
ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
ALLOWED_HEADERS=Content-Type,Authorization
ALLOWED_CREDENTIALS=true

# AIサービス
OLLAMA_URL=http://host.docker.internal:11434
OLLAMA_MODEL=gemma3:1b
EMBEDDING_MODEL=nomic-embed-text
SEARXNG_PUBLIC_URL=https://your-domain.com/search/

# 通知（オプション）
SLACK_WEBHOOK_URL=
EOF

chown -R $DEPLOY_USER:$DEPLOY_USER $DEPLOY_PATH
log_info "デプロイディレクトリ作成完了: $DEPLOY_PATH"

# 11. MySQL設定チューニング（メモリ節約）
log_info "MySQL設定をチューニングしています..."
mkdir -p /etc/mysql/conf.d/
cat > /etc/mysql/conf.d/tenhub.cnf << 'EOF'
[mysqld]
# メモリ節約設定（VPS 2GB向け）
innodb_buffer_pool_size = 256M
innodb_log_file_size = 64M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# 接続設定
max_connections = 50
max_connect_errors = 100

# キャッシュ
table_open_cache = 200
tmp_table_size = 16M
max_heap_table_size = 16M

# ログ
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2

# 文字セット
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci
EOF
log_info "MySQLチューニング完了"

# 12. 監視スクリプト設置
log_info "監視スクリプトを設置しています..."
cat > /usr/local/bin/tenhub-health.sh << 'EOF'
#!/bin/bash
# TenHub ヘルスチェック

ENDPOINT="http://localhost:3000"
DISCORD_WEBHOOK_URL=${DISCORD_WEBHOOK_URL:-}
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-}

check_service() {
    local name=$1
    local url=$2
    if ! curl -sf "$url" > /dev/null; then
        send_alert "$name is down!"
        return 1
    fi
    return 0
}

send_alert() {
    local message="$1"
    echo "[$(date)] $message" >> /var/log/tenhub-health.log

    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        curl -s -X POST "$SLACK_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"text\": \"❌ TenHub Alert: $message\"}"
    fi

    if [ -n "$DISCORD_WEBHOOK_URL" ]; then
        curl -s -X POST "$DISCORD_WEBHOOK_URL" \
            -H 'Content-Type: application/json' \
            -d "{\"content\": \"❌ TenHub Alert: $message\"}"
    fi
}

# メインチェック
check_service "Frontend" "$ENDPOINT" || exit 1
check_service "Gateway" "http://localhost:8080/health" || exit 1
EOF

chmod +x /usr/local/bin/tenhub-health.sh
log_info "監視スクリプト設置完了"

# 13. Cron設定（バックアップ・ヘルスチェック）
log_info "Cronジョブを設定しています..."
cat > /etc/cron.d/tenhub << EOF
# TenHub メンテナンスジョブ

# ヘルスチェック（5分おき）
*/5 * * * * root /usr/local/bin/tenhub-health.sh

# バックアップ（毎日3:00）
0 3 * * * root $DEPLOY_PATH/scripts/backup.sh

# Docker clean（毎週日曜4:00）
0 4 * * 0 root docker system prune -f
EOF

log_info "Cronジョブ設定完了"

# 14. まとめ情報
log_info "======================================"
log_info "初期設定完了！"
log_info "======================================"
echo ""
log_info "次のステップ:"
echo "  1. SSHポート: $SSH_PORT で接続してください"
echo "  2. $DEPLOY_USER ユーザーで docker コマンドが使えます"
echo "  3. デプロイパス: $DEPLOY_PATH"
echo "  4. .env.production を編集してください:"
echo "     vim $DEPLOY_PATH/.env.production"
echo ""
log_warn "重要: このターミナルを閉じる前に、新しいポートでSSH接続できることを確認してください！"
echo ""
log_info "新しい接続コマンド:"
echo "  ssh -p $SSH_PORT $DEPLOY_USER@$(curl -s ifconfig.me)"
echo ""
