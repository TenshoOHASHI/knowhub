.PHONY: proto nix-shell nix-zsh nix-proto help

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

help:
	@echo "TenHub root commands:"
	@echo ""
	@echo "=== Protocol Buffers ==="
	@echo "  make proto       - Generate Protocol Buffers"
	@echo ""
	@echo "=== Nix ==="
	@echo "  make nix-shell   - Enter Nix development shell"
	@echo "  make nix-zsh     - Enter Nix Zsh shell"
	@echo "  make nix-proto   - Generate proto files inside Nix environment"
	@echo ""
	@echo "Deployment commands are managed in deploy/Makefile."
	@echo "Example:"
	@echo "  make -C deploy help"
