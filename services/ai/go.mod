module github.com/TenshoOHASHI/knowhub/services/ai

go 1.25.4

replace github.com/TenshoOHASHI/knowhub/proto/wiki => ../../proto/wiki

replace github.com/TenshoOHASHI/knowhub/proto/ai => ../../proto/ai

replace github.com/TenshoOHASHI/knowhub/pkg => ../pkg

require (
	github.com/TenshoOHASHI/knowhub/pkg v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/ai v0.0.0-00010101000000-000000000000
	github.com/TenshoOHASHI/knowhub/proto/wiki v0.0.0-00010101000000-000000000000
	github.com/go-shiori/go-readability v0.0.0-20251205110129-5db1dc9836f0
	github.com/joho/godotenv v1.5.1
	google.golang.org/grpc v1.80.0
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/araddon/dateparse v0.0.0-20210429162001-6b43995a97de // indirect
	github.com/go-shiori/dom v0.0.0-20230515143342-73569d674e1c // indirect
	github.com/gogs/chardet v0.0.0-20211120154057-b7413eaefb8f // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120221211-b8f7ae30c516 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
