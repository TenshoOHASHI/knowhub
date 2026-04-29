module github.com/TenshoOHASHI/knowhub/services/gateway

go 1.25.4

require (
	github.com/TenshoOHASHI/knowhub/pkg v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/ai v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/auth v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/profile v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/wiki v0.0.0
	github.com/joho/godotenv v1.5.1
	github.com/swaggo/http-swagger v1.3.4
	github.com/swaggo/swag v1.16.6
	google.golang.org/grpc v1.80.0
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/swaggo/files v0.0.0-20220610200504-28940afbdbfe // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/TenshoOHASHI/knowhub/proto/wiki => ../../proto/wiki

replace github.com/TenshoOHASHI/knowhub/proto/auth => ../../proto/auth

replace github.com/TenshoOHASHI/knowhub/proto/profile => ../../proto/profile

replace github.com/TenshoOHASHI/knowhub/proto/ai => ../../proto/ai

replace github.com/TenshoOHASHI/knowhub/pkg => ../pkg
