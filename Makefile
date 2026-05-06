.PHONY: proto nix-shell nix-zsh nix-prodto prod-config prod-build prod-up prod-down prod-logs deploy-vps setup-vps backup health \
	deploy deploy-quick setup setup-check \
	dev dev-down dev-logs dev-restart test lint clean help

proto:
		protoc \
				--go_out=. \
				--go_opt=paths=source_relative \
				--go-grpc_out=. \
				--go-grpc_opt=paths=source_relative \
				proto/wiki/wiki.proto
		protoc \
				--go_out=. \
				--go_opt=paths=source_relative \
				--go-grpc_out=. \
				--go-grpc_opt=paths=source_relative \
				proto/auth/auth.proto

		protoc \
				--go_out=. \
				--go_opt=paths=source_relative \
				--go-grpc_out=. \
				--go-grpc_opt=paths=source_relative \
				proto/profile/profile.proto

		protoc \
				--go_out=. \
				--go_opt=paths=source_relative \
				--go-grpc_out=. \
				--go-grpc_opt=paths=source_relative \
				proto/ai/ai.proto

nix-shell:
		nix develop

nix-zsh:
		nix develop -c zsh -l

nix-proto:
		nix develop -c make proto

prod-config:
		docker compose -f docker-compose.prod.yml --env-file .env.production.example config

prod-build:
		docker compose -f docker-compose.prod.yml --env-file .env.production build

prod-up:
		docker compose -f docker-compose.prod.yml --env-file .env.production up -d --build

prod-down:
		docker compose -f docker-compose.prod.yml --env-file .env.production down

prod-logs:
		docker compose -f docker-compose.prod.yml --env-file .env.production logs -f --tail=200

# VPSデプロイ関連コマンド
VPS_HOST ?= $(shell echo $$VPS_HOST)
VPS_USER ?= $(shell echo $$VPS_USER)
DEPLOY_PATH ?= /opt/tenhub

deploy-vps:
	@echo "Deploying to VPS: $(VPS_USER)@$(VPS_HOST)"
	@scp -r docker-compose.prod.yml deploy $(VPS_USER)@$(VPS_HOST):$(DEPLOY_PATH)/
	@ssh $(VPS_USER)@$(VPS_HOST) "cd $(DEPLOY_PATH) && docker compose -f docker-compose.prod.yml pull && docker compose -f docker-compose.prod.yml up -d"

setup-vps:
	@echo "The legacy setup script is disabled."
	@echo "Read doc/lighthouse-setup.md and run the setup step by step."

backup:
	@echo "Running backup on VPS: $(VPS_USER)@$(VPS_HOST)"
	@ssh $(VPS_USER)@$(VPS_HOST) 'sudo bash -s' < deploy/scripts/backup.sh

health:
	@echo "Checking VPS health: $(VPS_USER)@$(VPS_HOST)"
	@ssh $(VPS_USER)@$(VPS_HOST) 'bash -s' < deploy/monitoring/health.sh

# ============================================
# シンプルデプロイ
# ============================================
deploy: deploy-quick ## VPS上でDockerイメージをpullして起動
	@echo "=== 完全デプロイ ==="

deploy-quick: ## クイックデプロイ（Dockerのみ）
	@echo "=== クイックデプロイ ==="
	@$(MAKE) -C deploy deploy-quick

setup: ## 初回構築ガイド表示
	@echo "=== 初回構築ガイド ==="
	@$(MAKE) -C deploy setup

setup-check: ## 設定チェック（ドライラン）
	@echo "=== 設定チェック（ドライラン）==="
	@$(MAKE) -C deploy setup-check

# ============================================
# 開発
# ============================================
dev: ## ローカル開発環境起動
	docker compose up -d
	@echo "開発環境を起動しました"
	@echo "Frontend: http://localhost:3000"
	@echo "Gateway API: http://localhost:8080"

dev-down: ## 開発環境停止
	docker compose down

dev-logs: ## 開発環境ログ
	docker compose logs -f

dev-restart: dev-down dev ## 開発環境再起動

# ============================================
# テスト・Lint
# ============================================
test: ## 全テスト実行
	@echo "テストを実行中..."
	cd services/auth && go test ./... -v || true
	cd services/wiki && go test ./... -v || true
	cd services/profile && go test ./... -v || true
	cd services/ai && go test ./... -v || true
	cd services/gateway && go test ./... -v || true

lint: ## Lint実行
	@echo "Lintを実行中..."
	cd services && gofmt -l .
	cd frontend && npm run lint || true

# ============================================
# クリーンアップ
# ============================================
clean: ## クリーンアップ
	docker compose down -v || true
	docker system prune -f
	@echo "クリーンアップ完了"

# ============================================
# ヘルプ
# ============================================
help: ## ヘルプ表示
	@echo "TenHub コマンド一覧:"
	@echo ""
	@echo "=== Protocol Buffers ==="
	@echo "  make proto         - Protocol Buffers生成"
	@echo ""
	@echo "=== Nix ==="
	@echo "  make nix-shell     - Nix開発シェル"
	@echo "  make nix-zsh       - Nix Zshシェル"
	@echo "  make nix-proto     - Nix環境でproto生成"
	@echo ""
	@echo "=== 本番環境（Docker Compose直接）==="
	@echo "  make prod-config   - 設定確認"
	@echo "  make prod-build    - ビルド"
	@echo "  make prod-up       - 起動"
	@echo "  make prod-down     - 停止"
	@echo "  make prod-logs     - ログ表示"
	@echo ""
	@echo "=== デプロイ（SSH + Docker Compose）==="
	@echo "  make deploy        - 完全デプロイ"
	@echo "  make deploy-quick  - クイックデプロイ"
	@echo "  make setup         - 初回構築ガイド表示"
	@echo "  make setup-check   - 設定チェック"
	@echo ""
	@echo "=== VPS操作（SSH直接）==="
	@echo "  make deploy-vps    - VPSへデプロイ"
	@echo "  make setup-vps     - VPS初期設定スクリプト実行"
	@echo "  make backup        - バックアップ実行"
	@echo "  make health        - ヘルスチェック"
	@echo ""
	@echo "=== 開発 ==="
	@echo "  make dev           - 開発環境起動"
	@echo "  make dev-down      - 開発環境停止"
	@echo "  make dev-logs      - ログ表示"
	@echo ""
	@echo "=== テスト・Lint ==="
	@echo "  make test          - テスト実行"
	@echo "  make lint          - Lint実行"
	@echo ""
	@echo "=== その他 ==="
	@echo "  make clean         - クリーンアップ"
	@echo "  make help          - このヘルプ表示"
