# knowhub - Workflow & Progress

## Working Rules（AIとの協業ルール）

- [x] AIはヒント＋コード例＋説明を提供する（コード例なしだと理解できない）
- [x] ユーザーが自分でコードを書く
- [x] AIがコードをレビューする
- [x] このスタイルでSPEC.mdの各フェーズを完了させる

## Design Decisions（設計上の決定事項）

### Model
- [x] Profile model と PortfolioItem model は別ファイルに分割（`profile.go` / `portfolioItem.go`）
- [x] `Status` フィールドは `PortfolioStatus` カスタム型 + 定数で定義（stringベース）
  - `StatusDeveloping` / `StatusCompleted`
- [x] `type enum struct{}` は不要なので削除

### Repository
- [x] 同じ `repository/` パッケージ内で別ファイルに分ける（Goの定番パターン）
  - 構造体名は `mysqlProfileRepository` / `mysqlPortfolioItemRepository` で衝突回避
- [x] Profile は1件のみ（自分用）→ `FindFirst` で取得（`FindAll`, `Delete` 不要）
- [x] 生SQLは標準 `database/sql` で統一（sqlx移行は将来検討）
- [x] Gateway の JWT Cookie / CORS / バリデーション は後のフェーズで実装

### Proto
- [x] `profile.proto` の `package` を `auth` → `profile` に修正する（TODO）

## Phase 6 Progress（Profile Service）

- [x] Proto definition & code generation
- [x] Profile model
- [x] PortfolioItem model
- [ ] PortfolioItem model の Status → PortfolioStatus カスタム型適用（TODO）
- [x] Profile repository
- [x] PortfolioItem repository（Save のSQL修正残し）
- [ ] gRPC handler
- [ ] main.go with service registration
- [ ] Migration のテーブル名修正（`articles` → `portfolio_items`）済み

## Phase 7 Progress（Frontend）

- [ ] Project setup
- [ ] Wiki page (article list + detail)
- [ ] Admin page (article editor)
- [ ] Profile page
- [ ] Portfolio page (developing / completed state)
- [ ] Dark mode

## Phase 8 Progress（Polish + Deploy）

- [ ] VPS setup (SSH, firewall)
- [ ] Nginx + Let's Encrypt
- [ ] Docker Compose production config
- [ ] GitHub Actions CI/CD
- [ ] slog structured logging
- [ ] Error response cleanup (hide internals)
- [ ] JWT Cookie / CORS / バリデーション（Gateway）

## Phase 9 Progress（AI Service）

- [ ] Article search with vector embeddings
- [ ] Article summarization
- [ ] Q&A based on wiki content
- [ ] Chat interface

## Phase 10 Progress（MCP Server）

- [ ] MCP Server implementation (Go)
- [ ] Tools: create_article, search_articles, list_articles
- [ ] Resources: article content access
- [ ] Integration with Claude Desktop / other AI assistants
