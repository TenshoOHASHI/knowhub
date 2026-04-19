# ディレクトリ作成
mkdir -p proto/auth
mkdir -p services/auth/internal/{model,repository,handler,jwt,config}
mkdir -p services/auth/migrations


# RPC

| RPC | Request | Response | 役割 |
| -- | -- | -- | -- |
| Register | username, email, password | User + Token | ユーザー登録 |
| Login | email, password | User + Token | ログイン |
| VerifyToken | token | User | トークン認証（他サービスから使う） |


# 型

```go

syntax = "proto3";

package auth;

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

option go_package = "github.com/TenshoOHASHI/knowhub/proto/auth";

```

# ディレクトリ作成

```bash
mkdir -p proto/auth
mkdir -p services/auth/internal/{model,repository,handler,jwt,config}
mkdir -p services/auth/migrations
```


# protoコード生成

```bash
protoc \
  --go_out=. \
  --go_opt=paths=source_relative \
  --go-grpc_out=. \
  --go-grpc_opt=paths=source_relative \
  proto/auth/auth.proto
```

# go mod init
```bash
cd services/auth
go mod init github.com/TenshoOHASHI/knowhub/services/auth
```

Makefileにも追加

```Makefile
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
```
