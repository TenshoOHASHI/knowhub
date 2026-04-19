# テスト
ツールをダウンロード
```bash
brew install grpcurl
```

# gRPC リフレクション（サービス一覧を返す機能）が有効にする必要がある
```go
 import (
      // ... 既存のimport ...
      "google.golang.org/grpc/reflection"    // 追加
  )

  func main() {
      // ...
      s := grpc.NewServer()
      pb.RegisterWikiServicesServer(s, wikiHandler)
      reflection.Register(s)    // この1行を追加

      // ...
  }
```

#  サービス一覧を確認
```bash
grpcurl -plaintext localhost:50052 list
```

# メソッド一覧を確認
```bash
grpcurl -plaintext localhost:50052 list wiki.WikiServices
```

# 記事を作成
```bash
grpcurl -plaintext -d '{"title": "Go入門", "content": "gRPCとは..."}' \
  localhost:50052 wiki.WikiServices/Create
```

# 記事一覧を取得
```bash
grpcurl -plaintext localhost:50052 wiki.WikiServices/List
```

# 記事一覧１件取得(Createで返ったIDを使う)
```bash
grpcurl -plaintext -d '{"id": "ここにIDを入れる"}' \
    localhost:50052 wiki.WikiServices/Get
```


# List（全記事取得）
grpcurl -plaintext localhost:50052 wiki.WikiServices/List


# Update（記事更新）
grpcurl -plaintext -d '{"id": "f8a1d7c6-8e80-4ff5-9ea4-63c5064694f8", "title": "Go中級"}' \
  localhost:50052 wiki.WikiServices/Update

# Delete（記事削除）
grpcurl -plaintext -d '{"id": "f8a1d7c6-8e80-4ff5-9ea4-63c5064694f8"}' \
  localhost:50052 wiki.WikiServices/Delete
