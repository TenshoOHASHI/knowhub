# ADR (Architecture Decision Records)

## ADR-001: Why Microservices?

### Context
個人サイトにマイクロサービスは過剰だが、学習目的で採用。

### Decision
4サービス構成（Gateway, Auth, Wiki, Profile）。

### Consequences
- (+) マイクロサービスの実践経験
- (+) gRPC通信の学習
- (-) ローカル開発の複雑さが増す
- (-) 運用オーバーヘッド

---

## ADR-002: Why CQRS in Wiki Service?

### Context
Wiki Serviceは読み取り頻度が高く、CQRSの学習に適している。

### Decision
Command/Query分離。Write → MySQL、Read → Redis。

### Consequences
- (+) CQRSパターンの理解
- (+) 読み取りパフォーマンスの向上
- (-) データの一貫性管理が複雑（Eventual Consistency）

---

## ADR-003: Why JWT RS256?

### Context
マイクロサービス間でステートレスな認証が必要。

### Decision
RS256（非対称鍵）を使用。Auth Serviceが秘密鍵で署名、他サービスは公開鍵で検証。

### Consequences
- (+) 各サービスがDBアクセスなしで検証可能
- (+) 鍵のローテーションが容易
- (-) 鍵管理の責任

---

## ADR-004: Proto-first Development

### Context
サービス間のインターフェースを先に定義し、契約駆動開発を行う。

### Decision
.protoファイルを先に書き、そこからGo/TSのコードを生成。

### Consequences
- (+) インターフェースが明確
- (+) 型安全な通信
- (-) Protoの学習コスト

---

## ADR-005: Why MySQL over PostgreSQL?

### Context
どちらもGoとの相性は良いが、開発者のMySQL経験を優先。

### Decision
MySQL 8を採用。Goドライバーは go-sql-driver/mysql を使用。

### Consequences
- (+) 馴染みのあるDBで学習効率が良い
- (+) 日本の求人でMySQL採用案件が多い
- (-) PostgreSQL特有の機能（JSONB等）は使えない
