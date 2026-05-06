# TenHub 本番デプロイ用の補助コマンドです。
# リポジトリ直下から実行する場合:
#   make -C deploy <target>
# deploy/ ディレクトリ内で実行する場合:
#   make <target>

# VPS上で本番ファイルを置くディレクトリです。
# `cd $(REMOTE_DEPLOY_PATH)` は、VPSに接続したあと
# `/opt/tenhub` に移動してから Docker Compose を実行する、という意味です。
REMOTE_DEPLOY_PATH ?= /opt/tenhub

# SSHハードニング後に使うSSHポートです。
# 初期構築では22番を使い、ハードニング後は ~/.ssh/config の
# `Host tenhub` 側もこのポートに変更します。
SSH_HARDENED_PORT ?= 22345

# ローカルで Docker Compose の構文確認をするときに使うファイルです。
# このMakefileは deploy/ 配下にあるため、`../` でリポジトリ直下を指します。
LOCAL_COMPOSE_FILE = ../docker-compose.prod.yml

.PHONY: help test setup setup-check deploy-quick deploy-ps deploy-logs logs-nginx logs-gateway health backup ssl setup-ssh-hardening

help:
	@echo "TenHub デプロイコマンド"
	@echo ""
	@echo "  make test                 - ローカルでcompose構文を確認"
	@echo "  make setup                - 初回構築ガイドを表示"
	@echo "  make setup-check          - testの別名"
	@echo "  make deploy-quick         - VPS上でimageをpullしてサービス起動"
	@echo "  make deploy-ps            - VPS上のcomposeサービス状態を表示"
	@echo "  make deploy-logs          - VPS上の全サービスログを表示"
	@echo "  make logs-nginx           - nginxログを表示"
	@echo "  make logs-gateway         - gatewayログを表示"
	@echo "  make health               - VPS上のヘルスチェックを実行"
	@echo "  make backup               - VPS上のバックアップを実行"
	@echo "  make ssl                  - Let's Encrypt証明書を取得してHTTPS化"
	@echo "  make setup-ssh-hardening  - SSHポート変更と認証強化"

test:
	@echo "=== Docker Compose config check ==="
	@docker compose -f $(LOCAL_COMPOSE_FILE) --env-file ../.env.production.example config > /dev/null
	@echo "OK"

setup-check: test

setup:
	@echo "まず doc/lighthouse-setup.md を読んでください。"
	@echo "今回の方針は、Tencent SSHキー、Firewall、Docker、ファイル配置を段階的に設定することです。"
	@echo "巨大なsetup shellを一発実行しない設計にしています。"

deploy-quick:
	# ssh tenhub:
	#   ~/.ssh/config の `Host tenhub` を使ってVPSへ接続します。
	# cd $(REMOTE_DEPLOY_PATH):
	#   VPS上の /opt/tenhub に移動します。
	# docker compose pull:
	#   GHCRから最新のDocker imageを取得します。
	# --env-file .env.production:
	#   image名や公開URLなど、Composeの変数展開に使う値を読み込みます。
	# docker compose up -d:
	#   コンテナをバックグラウンドで起動、または更新します。
	@ssh tenhub "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production pull && docker compose -f docker-compose.prod.yml --env-file .env.production up -d"
	@$(MAKE) deploy-ps

deploy-ps:
	# VPS上で起動している本番サービス一覧を表示します。
	@ssh tenhub "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production ps"

deploy-logs:
	# 本番サービス全体のログを表示します。
	# --tail=100 は、最初に表示する過去ログを最新100行に制限します。
	# -f は、新しいログをリアルタイムで追い続ける指定です。
	@ssh tenhub "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production logs -f --tail=100"

logs-nginx:
	# docker compose ps -q nginx で nginx のコンテナIDだけを取得します。
	# そのIDを docker logs に渡して、nginxだけのログを表示します。
	@ssh tenhub 'docker logs $$(docker compose -f $(REMOTE_DEPLOY_PATH)/docker-compose.prod.yml --env-file $(REMOTE_DEPLOY_PATH)/.env.production ps -q nginx) --tail=100 -f'

logs-gateway:
	# logs-nginx と同じ考え方で、Go API gateway のログだけを表示します。
	@ssh tenhub 'docker logs $$(docker compose -f $(REMOTE_DEPLOY_PATH)/docker-compose.prod.yml --env-file $(REMOTE_DEPLOY_PATH)/.env.production ps -q gateway) --tail=100 -f'

health:
	@ssh tenhub "$(REMOTE_DEPLOY_PATH)/scripts/health.sh"

backup:
	@ssh tenhub "$(REMOTE_DEPLOY_PATH)/scripts/backup.sh"

ssl:
	@echo "Domain name:"
	@read DOMAIN; \
	ssh tenhub "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production stop nginx || true && sudo certbot certonly --standalone -d $$DOMAIN --email admin@$$DOMAIN --agree-tos --non-interactive && (grep -q '^SERVER_NAME=' .env.production && sed -i 's/^SERVER_NAME=.*/SERVER_NAME=$$DOMAIN/' .env.production || echo SERVER_NAME=$$DOMAIN >> .env.production) && (grep -q '^PUBLIC_ORIGIN=' .env.production && sed -i 's#^PUBLIC_ORIGIN=.*#PUBLIC_ORIGIN=https://$$DOMAIN#' .env.production || echo PUBLIC_ORIGIN=https://$$DOMAIN >> .env.production) && (grep -q '^SSL_ENABLED=' .env.production && sed -i 's/^SSL_ENABLED=.*/SSL_ENABLED=true/' .env.production || echo SSL_ENABLED=true >> .env.production) && docker compose -f docker-compose.prod.yml --env-file .env.production up -d nginx"

setup-ssh-hardening:
	@echo "この処理ではSSHポートを $(SSH_HARDENED_PORT) に変更します。"
	@echo "実行前に Tencent Cloud Firewall で $(SSH_HARDENED_PORT)/tcp を許可してください。"
	@printf "続行しますか？ (y/N) "; read confirm; [ "$$confirm" = "y" ] || exit 1
	@ssh tenhub "sudo ufw allow $(SSH_HARDENED_PORT)/tcp && sudo sed -i 's/^#\\?Port .*/Port $(SSH_HARDENED_PORT)/' /etc/ssh/sshd_config && sudo sed -i 's/^#\\?PasswordAuthentication .*/PasswordAuthentication no/' /etc/ssh/sshd_config && sudo sed -i 's/^#\\?PermitRootLogin .*/PermitRootLogin no/' /etc/ssh/sshd_config && sudo systemctl restart ssh"
