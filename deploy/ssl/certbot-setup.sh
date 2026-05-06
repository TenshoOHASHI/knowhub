#!/bin/bash
set -e

# Usage:
#   sudo DEPLOY_PATH=/opt/tenhub ./certbot-setup.sh example.com admin@example.com

DOMAIN="$1"
EMAIL="${2:-admin@$DOMAIN}"
DEPLOY_PATH="${DEPLOY_PATH:-/opt/tenhub}"
COMPOSE_FILE="$DEPLOY_PATH/docker-compose.prod.yml"
ENV_FILE="$DEPLOY_PATH/.env.production"

if [ -z "$DOMAIN" ]; then
    echo "Usage: $0 example.com admin@example.com"
    exit 1
fi

echo "[1/5] Stop nginx container"
docker compose -f "$COMPOSE_FILE" stop nginx 2>/dev/null || true

echo "[2/5] Issue Let's Encrypt certificate"
certbot certonly --standalone \
    --non-interactive \
    --agree-tos \
    --email "$EMAIL" \
    -d "$DOMAIN" \
    --http-01-port=80

echo "[3/5] Ensure DH params"
if [ ! -f /etc/letsencrypt/ssl-dhparams.pem ]; then
    openssl dhparam -out /etc/letsencrypt/ssl-dhparams.pem 2048
fi

echo "[4/5] Update .env.production"
touch "$ENV_FILE"
grep -q '^SERVER_NAME=' "$ENV_FILE" && sed -i "s/^SERVER_NAME=.*/SERVER_NAME=$DOMAIN/" "$ENV_FILE" || echo "SERVER_NAME=$DOMAIN" >> "$ENV_FILE"
grep -q '^PUBLIC_ORIGIN=' "$ENV_FILE" && sed -i "s#^PUBLIC_ORIGIN=.*#PUBLIC_ORIGIN=https://$DOMAIN#" "$ENV_FILE" || echo "PUBLIC_ORIGIN=https://$DOMAIN" >> "$ENV_FILE"
grep -q '^SSL_ENABLED=' "$ENV_FILE" && sed -i 's/^SSL_ENABLED=.*/SSL_ENABLED=true/' "$ENV_FILE" || echo "SSL_ENABLED=true" >> "$ENV_FILE"

echo "[5/5] Start nginx"
docker compose -f "$COMPOSE_FILE" up -d nginx

echo "Done: https://$DOMAIN"
