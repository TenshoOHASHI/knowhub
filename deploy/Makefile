# TenHub 本番デプロイ用の補助コマンドです。
# リポジトリ直下から実行する場合:
#   make -C deploy <target>
# deploy/ ディレクトリ内で実行する場合:
#   make <target>

# GitHubに上げないローカル専用の環境変数ファイルです。
# make -C deploy で実行する場合、作業ディレクトリは deploy/ になるため、
# deploy/.env.local が読み込まれます。
LOCAL_ENV_FILE ?= .env.local

# .env.local が存在する場合だけ読み込みます。
# 存在しない場合は何もしません。
ifneq (,$(wildcard $(LOCAL_ENV_FILE)))
include $(LOCAL_ENV_FILE)
export
endif

# SSH接続先Host名です。
# セキュリティのため、Makefileには固定値を書きません。
# deploy/.env.local または実行時引数で指定してください。
SSH_HOST ?=

# VPS上で本番ファイルを置くディレクトリです。
# セキュリティのため、Makefileには本番パスを直接書きません。
REMOTE_DEPLOY_PATH ?=

# SSHハードニング後に使うSSHポートです。
# セキュリティのため、Makefileには本番ポート番号を直接書きません。
SSH_HARDENED_PORT ?=
# ローカルで Docker Compose の構文確認をするときに使うファイルです。
# このMakefileは deploy/ 配下にあるため、`../` でリポジトリ直下を指します。
LOCAL_COMPOSE_FILE = ../docker-compose.prod.yml

.PHONY: help test setup setup-check deploy-quick deploy-ps deploy-logs logs-nginx logs-gateway migrate-status migrate-up migrate-down health backup ssl setup-ssh-hardening require-remote-deploy-path require-ssh-hardened-port

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
	@echo "  make migrate-status       - goose migration状態を確認"
	@echo "  make migrate-up           - 未実行migrationを適用"
	@echo "  make migrate-down         - 直近migrationを1つ戻す"
	@echo "  make health               - VPS上のヘルスチェックを実行"
	@echo "  make backup               - VPS上のバックアップを実行"
	@echo "  make ssl                  - Let's Encrypt証明書を取得してHTTPS化"
	@echo "  make setup-ssh-hardening  - SSHポート変更と認証強化"
	@echo ""
	@echo "必須変数:"
	@echo "  REMOTE_DEPLOY_PATH        - VPS上の本番配置ディレクトリ"
	@echo "  SSH_HARDENED_PORT         - setup-ssh-hardening 実行時のみ必要"
	@echo ""
	@echo "任意変数:"
	@echo "  SSH_HOST                  - SSH接続先Host名"

require-remote-deploy-path:
	@test -n "$(REMOTE_DEPLOY_PATH)" || (echo "ERROR: REMOTE_DEPLOY_PATH is required"; exit 1)

require-ssh-hardened-port:
	@test -n "$(SSH_HARDENED_PORT)" || (echo "ERROR: SSH_HARDENED_PORT is required"; exit 1)

test:
	@echo "=== Docker Compose config check ==="
	@docker compose -f $(LOCAL_COMPOSE_FILE) --env-file ../.env.production.example config > /dev/null
	@echo "OK"

setup-check: test

setup:
	@echo "まず doc/lighthouse-setup.md を読んでください。"
	@echo "今回の方針は、SSHキー、Firewall、Docker、ファイル配置を段階的に設定することです。"
	@echo "巨大なsetup shellを一発実行しない設計にしています。"

deploy-quick: require-remote-deploy-path
	# ssh $(SSH_HOST):
	#   ~/.ssh/config の Host を使ってVPSへ接続します。
	# cd $(REMOTE_DEPLOY_PATH):
	#   VPS上の本番配置ディレクトリに移動します。
	# docker compose pull:
	#   Registryから最新のDocker imageを取得します。
	# --env-file .env.production:
	#   image名や公開URLなど、Composeの変数展開に使う値を読み込みます。
	# docker compose up -d:
	#   コンテナをバックグラウンドで起動、または更新します。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production pull"
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production up -d"
	@$(MAKE) deploy-ps REMOTE_DEPLOY_PATH="$(REMOTE_DEPLOY_PATH)" SSH_HOST="$(SSH_HOST)"

deploy-ps: require-remote-deploy-path
	# VPS上で起動している本番サービス一覧を表示します。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production ps"

deploy-logs: require-remote-deploy-path
	# 本番サービス全体のログを表示します。
	# --tail=100 は、最初に表示する過去ログを最新100行に制限します。
	# -f は、新しいログをリアルタイムで追い続ける指定です。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production logs -f --tail=100"

logs-nginx: require-remote-deploy-path
	# docker compose ps -q nginx で nginx のコンテナIDだけを取得します。
	# そのIDを docker logs に渡して、nginxだけのログを表示します。
	@ssh $(SSH_HOST) 'docker logs $$(docker compose -f $(REMOTE_DEPLOY_PATH)/docker-compose.prod.yml --env-file $(REMOTE_DEPLOY_PATH)/.env.production ps -q nginx) --tail=100 -f'

logs-gateway: require-remote-deploy-path
	# logs-nginx と同じ考え方で、Go API gateway のログだけを表示します。
	@ssh $(SSH_HOST) 'docker logs $$(docker compose -f $(REMOTE_DEPLOY_PATH)/docker-compose.prod.yml --env-file $(REMOTE_DEPLOY_PATH)/.env.production ps -q gateway) --tail=100 -f'

migrate-status: require-remote-deploy-path
	# goose_db_version を見て、適用済み/Pendingのmigrationを表示します。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production run --rm migrator status"

migrate-up: require-remote-deploy-path
	# 未実行のmigrationを順番に適用します。deploy前に手動で実行します。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production run --rm migrator up"

migrate-down: require-remote-deploy-path
	# 直近のmigrationを1つ戻します。データ削除を伴う可能性があるため慎重に使います。
	@ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production run --rm migrator down"

health: require-remote-deploy-path
	@ssh $(SSH_HOST) "$(REMOTE_DEPLOY_PATH)/scripts/health.sh"

backup: require-remote-deploy-path
	@ssh $(SSH_HOST) "$(REMOTE_DEPLOY_PATH)/scripts/backup.sh"

ssl: require-remote-deploy-path
	@echo "Domain name:"
	@read DOMAIN; \
	ssh $(SSH_HOST) "cd $(REMOTE_DEPLOY_PATH) && docker compose -f docker-compose.prod.yml --env-file .env.production stop nginx || true && sudo certbot certonly --standalone -d $$DOMAIN --email admin@$$DOMAIN --agree-tos --non-interactive && (grep -q '^SERVER_NAME=' .env.production && sed -i 's/^SERVER_NAME=.*/SERVER_NAME=$$DOMAIN/' .env.production || echo SERVER_NAME=$$DOMAIN >> .env.production) && (grep -q '^PUBLIC_ORIGIN=' .env.production && sed -i 's#^PUBLIC_ORIGIN=.*#PUBLIC_ORIGIN=https://$$DOMAIN#' .env.production || echo PUBLIC_ORIGIN=https://$$DOMAIN >> .env.production) && (grep -q '^SSL_ENABLED=' .env.production && sed -i 's/^SSL_ENABLED=.*/SSL_ENABLED=true/' .env.production || echo SSL_ENABLED=true >> .env.production) && docker compose -f docker-compose.prod.yml --env-file .env.production up -d nginx"

setup-ssh-hardening: require-ssh-hardened-port
	@echo "この処理ではSSHポートを $(SSH_HARDENED_PORT) に変更します。"
	@echo "実行前に Cloud Firewall で $(SSH_HARDENED_PORT)/tcp を許可してください。"
	@printf "続行しますか？ (y/N) "; read confirm; [ "$$confirm" = "y" ] || exit 1
	@ssh $(SSH_HOST) "sudo ufw allow $(SSH_HARDENED_PORT)/tcp && sudo sed -i 's/^#\\?Port .*/Port $(SSH_HARDENED_PORT)/' /etc/ssh/sshd_config && sudo sed -i 's/^#\\?PasswordAuthentication .*/PasswordAuthentication no/' /etc/ssh/sshd_config && sudo sed -i 's/^#\\?PermitRootLogin .*/PermitRootLogin no/' /etc/ssh/sshd_config && sudo systemctl restart ssh"
