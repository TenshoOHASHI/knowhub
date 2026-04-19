module github.com/TenshoOHASHI/knowhub/services/wiki

go 1.25.4

require (
	github.com/TenshoOHASHI/knowhub/proto/wiki v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.9.3
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.80.0
	google.golang.org/protobuf v1.36.11
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
)

// ローカルのprotoパッケージを参照
replace github.com/TenshoOHASHI/knowhub/proto/wiki => ../../proto/wiki
