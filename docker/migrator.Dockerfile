# ==========================================
# Goose migrator image
# ==========================================
# DB migration だけを実行する小さな Go バイナリを作ります。
# アプリ本体とは分けておくことで、VPS では必要な時だけ
# `docker compose run --rm migrator up`
# のように実行できます。

# 使用する Go のバージョン。
# docker build 時に --build-arg GO_VERSION=1.25 のように上書きできます。
# この値は `FROM golang:${GO_VERSION}-alpine` で使われます。
ARG GO_VERSION=1.25

# goose は複数の DB driver に対応しています。
# このプロジェクトでは MySQL だけを使うため、
# 不要な driver を build tag で無効化してバイナリを軽くします。
# no_postgres などは「PostgreSQL driverをbuild対象から外す」という意味です。
ARG GOOSE_BUILD_TAGS="no_clickhouse no_libsql no_mssql no_postgres no_sqlite3 no_vertica no_ydb"

# ==========================================
# Build stage
# ==========================================
# Go のビルド用イメージを使います。
# alpine ベースなので比較的軽量です。
# AS builder:
#   このstageに builder という名前を付けます。
#   後段の `COPY --from=builder ...` で参照します。
FROM golang:${GO_VERSION}-alpine AS builder

# コンテナ内の作業ディレクトリを設定します。
# 以降の COPY / RUN は、この /src を基準に実行されます。
WORKDIR /src

# 先に go.mod / go.sum だけコピーします。
# こうすることで、依存関係のダウンロード結果が Docker layer cache に乗りやすくなります。
# ローカル:
#   services/migrator/go.mod
#   services/migrator/go.sum
#
# コンテナ内:
#   /src/go.mod
#   /src/go.sum
COPY services/migrator/go.mod services/migrator/go.sum* ./

# Go module の依存関係をダウンロードします。
# go.mod / go.sum が変わらなければ、このlayerは再利用されやすくなります。
RUN go mod download

# migrator のソースコード全体をコピーします。
# ローカル:
#   services/migrator/
#
# コンテナ内:
#   /src/
#
# 例:
#   services/migrator/main.go -> /src/main.go
#   services/migrator/go.mod  -> /src/go.mod
COPY services/migrator ./

# migrator バイナリをビルドします。
#
# CGO_ENABLED=0:
#   C ライブラリに依存しない静的寄りのバイナリを作ります。
#
# GOOS=linux:
#   Linux コンテナ内で動くバイナリを作ります。
#
# -tags="${GOOSE_BUILD_TAGS}":
#   goose の不要な DB driver を除外します。
#
# -trimpath:
#   ビルド時のローカルパス情報をバイナリから削ります。
#
# -ldflags="-s -w":
#   デバッグ情報などを削ってバイナリサイズを小さくします。
#
# -o /out/migrator:
#   出力先を /out/migrator にします。
# 最終的に runtime stage へコピーするのは、この /out/migrator だけです。
RUN CGO_ENABLED=0 GOOS=linux go build \
    -tags="${GOOSE_BUILD_TAGS}" \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/migrator .

# ==========================================
# Runtime stage
# ==========================================
# 実行用の最小イメージです。
# Go のビルド環境は含めず、完成したバイナリだけを入れます。
# これにより、本番imageを軽くし、不要なbuild toolを含めないようにします。
FROM alpine:3.22

# ca-certificates:
#   TLS 接続が必要な場合に証明書を使えるようにします。
#
# tzdata:
#   タイムゾーン情報を使えるようにします。
RUN apk add --no-cache ca-certificates tzdata

# 実行時の作業ディレクトリを設定します。
WORKDIR /app

# builder stage で作った migrator バイナリを runtime image にコピーします。
# /usr/local/bin に置くことで PATH から実行できます。
# コピー元:
#   builder stage の /out/migrator
#
# コピー先:
#   runtime stage の /usr/local/bin/tenhub-migrator
COPY --from=builder /out/migrator /usr/local/bin/tenhub-migrator

# goose が読む SQL migration ファイルを image に同梱します。
# これにより、VPS 側に Go や goose CLI を直接インストールする必要がありません。
# ローカル:
#   deploy/migrations/
#
# コンテナ内:
#   /migrations/
#
# 例:
#   deploy/migrations/001_initial_schema.sql
#     -> /migrations/001_initial_schema.sql
COPY deploy/migrations /migrations

# migrator プログラムが参照する migration directory を環境変数で指定します。
# services/migrator/main.go は MIGRATION_DIR を読み、未設定なら /migrations を使います。
ENV MIGRATION_DIR=/migrations

# コンテナ起動時に実行するコマンド本体です。
# 例:
#   docker compose run --rm migrator
#   -> tenhub-migrator が実行されます。
ENTRYPOINT ["tenhub-migrator"]

# デフォルト引数です。
# 引数を指定しない場合は `tenhub-migrator status` が実行されます。
#
# 例:
#   docker compose run --rm migrator
#   -> tenhub-migrator status
#
# 例:
#   docker compose run --rm migrator up
#   -> tenhub-migrator up
CMD ["status"]
