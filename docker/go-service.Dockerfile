# ==========================================
# 1. ビルド用ステージ (builder)
# ==========================================
# Goのバージョンを変数化
ARG GO_VERSION=1.25

# ビルド環境用のイメージ
FROM golang:${GO_VERSION}-alpine AS builder

# 対象サービス名を受け取る
ARG SERVICE
WORKDIR /src

# 必要なツールのインストール
RUN apk add --no-cache git

# 共通ファイルと対象サービスのソースをコピー
COPY proto ./proto
COPY services/pkg ./services/pkg
COPY services/${SERVICE} ./services/${SERVICE}

# 依存関係の解決
WORKDIR /src/services/${SERVICE}
RUN go mod download

# 【修正箇所】Goバイナリのビルド
# 1行で記述し、出力先を /out/service に固定します
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service .

# ==========================================
# 2. 実行用ステージ (Final Image)
# ==========================================
# 実行用の軽量OS
FROM alpine:3.22

# 実行時用の環境変数
ARG SERVICE
ENV SERVICE=${SERVICE}

# 証明書とタイムゾーンデータの追加
RUN apk add --no-cache ca-certificates tzdata

# 実行ディレクトリの作成
WORKDIR /app/services/${SERVICE}

# builderからバイナリだけを盗んでくる（コピーする）
COPY --from=builder /out/service /usr/local/bin/tenhub-service

# ポート開放
EXPOSE 50051 50052 50053 50054 8080

# docker compose up時に、アプリ起動
CMD ["tenhub-service"]
