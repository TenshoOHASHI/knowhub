#!/bin/bash
# Let's Encrypt証明書を取得し、TenHub本番envをHTTPS用に更新するスクリプトです。
# ドメイン取得後にVPS上で実行します。

set -euo pipefail
# set -e:
#   途中の失敗で終了します。
# set -u:
#   未定義変数を使ったら終了します。
# set -o pipefail:
#   pipeを使う処理で途中の失敗も検知します。

# Usage:
#   sudo DEPLOY_PATH=/opt/tenhub ./certbot-setup.sh example.com admin@example.com
#
# 注意:
#   このスクリプトは証明書取得後に .env.production を HTTPS 用へ切り替えます。
#   HTTP 検証中は COOKIE_SECURE=false ですが、HTTPS 化後は COOKIE_SECURE=true にします。

# $1:
#   証明書を取得したいドメインです。
#   例: tenhub.example.com
DOMAIN="${1:-}"

# $2:
#   Let's Encryptの通知用メールアドレスです。
#   未指定なら admin@<DOMAIN> を使います。
EMAIL="${2:-admin@$DOMAIN}"

# DEPLOY_PATH:
#   VPS上のTenHub配置先です。
#   実行時に DEPLOY_PATH=/path のように上書きできます。
DEPLOY_PATH="${DEPLOY_PATH:-/opt/tenhub}"
COMPOSE_FILE="$DEPLOY_PATH/docker-compose.prod.yml"
ENV_FILE="$DEPLOY_PATH/.env.production"

if [ -z "$DOMAIN" ]; then
    echo "Usage: $0 example.com admin@example.com"
    exit 1
fi

echo "[1/6] Stop nginx container"
# certbot --standalone は80番ポートを一時的に使います。
# nginxが80番を使っているため、証明書取得中だけnginxコンテナを止めます。
# `|| true` は、nginxが未起動でもスクリプトを止めないためです。
docker compose -f "$COMPOSE_FILE" stop nginx 2>/dev/null || true

echo "[2/6] Issue Let's Encrypt certificate"
# certbot certonly --standalone:
#   nginxとは別にcertbot自身が一時Webサーバーを立てて認証します。
# --non-interactive:
#   対話入力なしで実行します。
# --agree-tos:
#   Let's Encryptの利用規約に同意します。
# --http-01-port=80:
#   HTTP-01認証で80番ポートを使います。
certbot certonly --standalone \
    --non-interactive \
    --agree-tos \
    --email "$EMAIL" \
    -d "$DOMAIN" \
    --http-01-port=80

echo "[3/6] Ensure DH params"
# ssl-dhparams.pem:
#   TLSの安全性を高めるためのDHパラメータです。
#   既に存在する場合は作り直しません。
if [ ! -f /etc/letsencrypt/ssl-dhparams.pem ]; then
    openssl dhparam -out /etc/letsencrypt/ssl-dhparams.pem 2048
fi

echo "[4/6] Update .env.production"
touch "$ENV_FILE"

# grep -q:
#   対象の設定行が既にあるかを確認します。表示はしません。
# sed -i:
#   既存行をファイル内で直接置換します。
# || echo ... >>:
#   既存行がなければ末尾に追記します。
#
# SERVER_NAME:
#   nginxが受け付けるドメイン名です。
grep -q '^SERVER_NAME=' "$ENV_FILE" && sed -i "s/^SERVER_NAME=.*/SERVER_NAME=$DOMAIN/" "$ENV_FILE" || echo "SERVER_NAME=$DOMAIN" >> "$ENV_FILE"

# PUBLIC_ORIGIN:
#   フロントエンドやCORSで使う公開URLです。
grep -q '^PUBLIC_ORIGIN=' "$ENV_FILE" && sed -i "s#^PUBLIC_ORIGIN=.*#PUBLIC_ORIGIN=https://$DOMAIN#" "$ENV_FILE" || echo "PUBLIC_ORIGIN=https://$DOMAIN" >> "$ENV_FILE"

# SSL_ENABLED:
#   nginx templateでHTTPS設定を有効化するためのフラグです。
grep -q '^SSL_ENABLED=' "$ENV_FILE" && sed -i 's/^SSL_ENABLED=.*/SSL_ENABLED=true/' "$ENV_FILE" || echo "SSL_ENABLED=true" >> "$ENV_FILE"

# COOKIE_SECURE:
#   Next.js Route Handler が Set-Cookie する時の Secure 属性です。
#   HTTPS化後は true にして、CookieをHTTPSでだけ送信します。
grep -q '^COOKIE_SECURE=' "$ENV_FILE" && sed -i 's/^COOKIE_SECURE=.*/COOKIE_SECURE=true/' "$ENV_FILE" || echo "COOKIE_SECURE=true" >> "$ENV_FILE"

# SEARXNG_PUBLIC_URL:
#   SearXNG自身が知る公開URLです。
#   HTTPS化後は http:// ではなく https:// に揃えます。
grep -q '^SEARXNG_PUBLIC_URL=' "$ENV_FILE" && sed -i "s#^SEARXNG_PUBLIC_URL=.*#SEARXNG_PUBLIC_URL=https://$DOMAIN/search/#" "$ENV_FILE" || echo "SEARXNG_PUBLIC_URL=https://$DOMAIN/search/" >> "$ENV_FILE"

echo "[5/6] Install certbot renewal hooks"
# standalone方式では、証明書更新時にも80番ポートが必要です。
# 通常はnginxコンテナが80番を使っているため、
# certbot renew の前にnginxを止め、更新後にnginxを戻すhookを配置します。
mkdir -p /etc/letsencrypt/renewal-hooks/pre /etc/letsencrypt/renewal-hooks/post

cat > /etc/letsencrypt/renewal-hooks/pre/tenhub-stop-nginx.sh <<EOF
#!/bin/sh
cd "$DEPLOY_PATH" || exit 0
docker compose -f docker-compose.prod.yml --env-file .env.production stop nginx || true
EOF

cat > /etc/letsencrypt/renewal-hooks/post/tenhub-start-nginx.sh <<EOF
#!/bin/sh
cd "$DEPLOY_PATH" || exit 0
docker compose -f docker-compose.prod.yml --env-file .env.production up -d nginx || true
EOF

chmod +x /etc/letsencrypt/renewal-hooks/pre/tenhub-stop-nginx.sh
chmod +x /etc/letsencrypt/renewal-hooks/post/tenhub-start-nginx.sh

echo "[6/6] Start nginx"
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d nginx

echo "Done: https://$DOMAIN"
