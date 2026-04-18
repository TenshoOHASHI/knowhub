# Goのプラグインをインストール
Step1: protoc-gen-go: proto → Goの構造体を生成するプラグイン

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Step2:  proto → gRPCのクライアント/サーバー コードを生成するプラグイン

```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

ポイント:
- go install は $GOPATH/bin（または $HOME/go/bin）にバイナリを置く
- このパスが $PATH に含まれていないと protoc がプラグインを見つけられない

Step3: パスを確認する

```bash
which protoc-gen-go
which protoc-gen-go-grpc
```

もしなければ

~/.zshrc または ~/.bashrc に追加
export PATH="$PATH:$(go env GOPATH)/bin"

# Go のコード生成

コマンドの構成
protoc \
  --go_out=. \　# Go構造体の出力先
  --go_out=paths=source_relative \　# protoと同じディレクトリ構造で出力、指定しな場合、option go_package直下にファイルが出力される
  --go-grpc_out=. \　# gRPCコード出力先
  --go-grpc_opt=paths=source_relative \
  proto/wiki/wiki.proto  # 対象のprotoファイル

各オプションの意味

--go_out=.
Go構造体（message → struct）を .（プロジェクトルート）基準で出力

--go-grpc_out=.
gRPCコード（service → interface）を出力

paths=source_relative
protoファイルの場所にそのまま出力（go_packageのパスを無視）
