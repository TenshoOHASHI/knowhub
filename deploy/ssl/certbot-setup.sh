#!/bin/bash
# Let's Encrypt SSL証明書取得・設定スクリプト
# 使い方: sudo ./certbot-setup.sh your-domain.com

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

# 引数チェック
if [ -z "$1" ]; then
    log_error "ドメインを指定してください"
    echo "使い方: $0 your-domain.com"
    exit 1
fi

DOMAIN=$1
EMAIL=${2:-"admin@$DOMAIN"}

log_info "SSL証明書設定を開始します: $DOMAIN"

# 1. Nginxが一時停止していることを確認
log_info "Nginxを一時停止します..."
docker stop tenhub-nginx-1 2>/dev/null || true
systemctl stop nginx 2>/dev/null || true

# 2. Certbotで証明書取得
log_info "Let's Encryptで証明書を取得します..."
certbot certonly --standalone \
    --non-interactive \
    --agree-tos \
    --email "$EMAIL" \
    -d "$DOMAIN" \
    --http-01-port=80

if [ $? -eq 0 ]; then
    log_info "証明書取得成功！"
else
    log_error "証明書取得失敗"
    exit 1
fi

# 3. 証明書パス
CERT_PATH="/etc/letsencrypt/live/$DOMAIN"
log_info "証明書パス: $CERT_PATH"
log_info "  証明書: $CERT_PATH/fullchain.pem"
log_info "  秘密鍵: $CERT_PATH/privkey.pem"

# 4. Diffie-Hellmanパラメータ生成（一回のみ）
if [ ! -f /etc/letsencrypt/ssl-dhparams.pem ]; then
    log_info "DHパラメータを生成しています（時間がかかります）..."
    openssl dhparam -out /etc/letsencrypt/ssl-dhparams.pem 2048
fi

# 5. Nginx SSLテンプレート更新
log_info "Nginx SSL設定を更新します..."

# deploy/nginx/templates/ssl.conf.template を作成
mkdir -p /opt/tenhub/deploy/nginx/templates

cat > /opt/tenhub/deploy/nginx/templates/ssl.conf.template << EOF
server {
    listen 80;
    server_name \${SERVER_NAME};

    # Let's Encrypt HTTP-01チャレンジ用
    location /.well-known/acme-challenge/ {
        root /var/www/certbot;
    }

    # HTTPSにリダイレクト
    location / {
        return 301 https://\$host\$request_uri;
    }
}

server {
    listen 443 ssl http2;
    server_name \${SERVER_NAME};

    # SSL証明書
    ssl_certificate /etc/letsencrypt/live/\${SERVER_NAME}/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/\${SERVER_NAME}/privkey.pem;
    ssl_trusted_certificate /etc/letsencrypt/live/\${SERVER_NAME}/chain.pem;

    # SSL設定（推奨）
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # DHパラメータ
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    client_max_body_size 20m;

    # ヘッダー
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    location / {
        proxy_pass http://frontend:3000;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
EOF

# 6. docker-compose.yml更新でNginxにボリュームマウント追加
log_info "docker-compose.ymlにSSLボリュームを追加してください:"
echo ""
echo "nginx:"
echo "  volumes:"
echo "    - ./deploy/nginx/templates:/etc/nginx/templates:ro"
echo "    - /etc/letsencrypt:/etc/letsencrypt:ro"
echo "  ports:"
echo "    - '80:80'"
echo "    - '443:443'"
echo ""

# 7. 自動更新設定
log_info "証明書自動更新cronジョブを設定します..."

# 更新スクリプト作成
cat > /usr/local/bin/certbot-renew.sh << 'EOF'
#!/bin/bash
# Let's Encrypt証明書自動更新スクリプト

certbot renew --quiet --deploy-hook "
    docker restart tenhub-nginx-1 2>/dev/null || true
    systemctl reload nginx 2>/dev/null || true
"
EOF

chmod +x /usr/local/bin/certbot-renew.sh

# Cronに追加（毎日3:30と15:30にチェック）
(crontab -l 2>/dev/null | grep -v "certbot-renew"; echo "30 3,15 * * * /usr/local/bin/certbot-renew.sh >> /var/log/certbot-renew.log 2>&1") | crontab -

log_info "自動更新設定完了（毎日3:30と15:30にチェック）"

# 8. 証明書情報表示
log_info "証明書情報:"
openssl x509 -in "$CERT_PATH/fullchain.pem" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After|DNS:)" || true

# 9. 次のステップ
log_info "======================================"
log_info "SSL設定完了！"
log_info "======================================"
echo ""
log_info "次のステップ:"
echo "  1. docker-compose.ymlを更新してSSLボリュームをマウント"
echo "  2. docker compose up -d --force-recreate nginx"
echo "  3. https://$DOMAIN にアクセスして確認"
echo ""
log_info "証明書有効期限チェック:"
echo "  certbot certificates"
echo ""
log_info "手動更新:"
echo "  certbot renew --dry-run"
echo ""
