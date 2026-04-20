# API Gateway

|機能 | 内容 |
| -- | -- |
|REST -> gRPC　変換 | HTTPリクエストをgRPCに変換して各サービスに転送 |
| JWT検証 | 保護されたエンドポイントでトークンをチェック |
| CORS | ブラウザからのリクエストを許可 |
|ルーティング | /api/articles -> Wiki, /api/auth -> Auth |


# エンドポイントの設計

Public (認証不要):
- POST /api/auth/register -> Auth Service / List
- POST /api/auth/Login -> Auth Service / Login

Protected (認証必要):
- GET /api/articles -> Wiki Service / List
- GET /api/articles/:id -> Wiki Service / Get
- POST /api/articles -> Wiki Service / Create
- PUT /api/articles/:id -> Wiki Service / Update
- Delete /api/articles/:id -> Wiki Service / Delete


# ディレクトリ構成

services/gateway/
  ├── main.go
  ├── internal/
  │   ├── handler/
  │   │   ├── auth_handler.go      ← REST → gRPC (Auth)
  │   │   └── wiki_handler.go      ← REST → gRPC (Wiki)
  │   ├── middleware/
  │   │   └── auth.go              ← JWT検証ミドルウェア
  │   └── config/
  │       └── config.go

#  必要なパッケージ
```bash
cd /Users/oohashitenshou/Desktop/my-portfolio/services/gateway
go get github.com/joho/godotenv
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

HTTPルーターには標準の net/http + Go 1.22 の http.NewServeMux を使用（外部パッケージ不要）


# API　テスト

## ポート確認
- lsof -i :8080

## 全記事を取得
```bash
curl http://localhost:8080/api/articles
```

## 特定の記事を取得
```bash
curl http://localhost:8080/api/articles/aba767a2-9373-44da-b195-1c76df150fcb
```

## 記事を作成
```bash
curl -X POST http://localhost:8080/api/articles -H "Content-Type: application/json" -d '{"title":"test","content":"Go入門"}'
```

## 記事を更新
```bash
curl -X PUT http://localhost:8080/api/articles/aba767a2-9373-44da-b195-1c76df150fcb -H "Content-Type: application/json" -d '{"title":"test","content":"Go中級"}'
```

## 特定の記事を削除
```bash
curl -X DELETE http://localhost:8080/api/articles/aba767a2-9373-44da-b195-1c76df150fcb
```


## ユーザーを作成
```bash
curl -X POST http://localhost:8080/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
```

## ログイン
```bash
curl -X POST http://localhost:8080/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```
