module github.com/TenshoOHASHI/knowhub/services/gateway

go 1.25.4

require (
	github.com/TenshoOHASHI/knowhub/proto/auth v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/wiki v0.0.0
	google.golang.org/grpc v1.80.0
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/TenshoOHASHI/knowhub/proto/wiki => ../../proto/wiki

replace github.com/TenshoOHASHI/knowhub/proto/auth => ../../proto/auth
