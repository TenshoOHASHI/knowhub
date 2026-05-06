.PHONY: proto nix-shell nix-zsh nix-prodto prod-config prod-build prod-up prod-down prod-logs deploy-vps setup-vps backup health

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
	@echo "Running VPS setup script on $(VPS_USER)@$(VPS_HOST)"
	@ssh $(VPS_USER)@$(VPS_HOST) 'bash -s' < deploy/scripts/setup-vps.sh

backup:
	@echo "Running backup on VPS: $(VPS_USER)@$(VPS_HOST)"
	@ssh $(VPS_USER)@$(VPS_HOST) 'sudo bash -s' < deploy/scripts/backup.sh

health:
	@echo "Checking VPS health: $(VPS_USER)@$(VPS_HOST)"
	@ssh $(VPS_USER)@$(VPS_HOST) 'bash -s' < deploy/monitoring/health.sh
