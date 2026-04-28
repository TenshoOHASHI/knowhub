target:
	command

.PHONY: proto

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
