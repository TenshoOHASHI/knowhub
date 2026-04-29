# knowhub - Product Specification

## Overview
技術ナレッジベースプラットフォーム。
自身の学習記録をWikiとして蓄積し、そのプロジェクト自体がポートフォリオとなる。
バックエンドはGo（マイクロサービス + gRPC + CQRS）、フロントエンドはNext.js。

## Concept
- **技術書を書く感覚**で学習内容をアウトプットする場所
- このプロジェクト自体が最大のポートフォリオ作品
- AI活用による知識検索・要約などの面白い機能を追加予定

## Target User
- 自分自身（学習記録の管理・検索）
- 採用担当者（技術力・継続力の可視化）
- AIアシスタント（MCP経由でナレッジにアクセス）

## Tech Stack

### Backend
- Language: Go 1.22+
- Architecture: Microservices
- Communication: gRPC (inter-service), REST (external)
- Auth: JWT (RS256)
- Patterns: CQRS (Wiki Service)
- AI: OpenAI API / Local LLM (Phase 9)
- MCP: Model Context Protocol Server (Phase 10)

### Frontend
- Framework: Next.js 14+ (App Router)
- Language: TypeScript
- Styling: Tailwind CSS

### Infrastructure
- Container: Docker + Docker Compose
- DB: MySQL 8 (Write), Redis 7 (Read Cache)
- Proto: Protocol Buffers 3
- Deploy: VPS (Hostinger等)
- CI/CD: GitHub Actions

## Services

| Service | Port | Role |
|---------|------|------|
| API Gateway | :8080 (REST) | JWT validation, routing, CORS |
| Auth Service | :50051 (gRPC) | Register, Login, JWT |
| Wiki Service | :50052 (gRPC) | Article CRUD (CQRS) |
| Profile Service | :50053 (gRPC) | Self-intro, portfolio items |
| AI Service | :50054 (gRPC) | Article search, summarization (Phase 9) |
| MCP Server | :5005 (stdio/SSE) | AI assistant integration (Phase 10) |

## Database Design

### MySQL (Auth Service)
- users table

### MySQL (Wiki Service)
- articles table

### MySQL (Profile Service)
- profiles table
- portfolio_items table

### Redis (Wiki Service - Read Model)
- article:{id} → cached article JSON
- articles:list → cached article list

## Development Phases

### Phase 1: Project Setup ✅
- [x] Create project directory
- [x] Docker Compose setup (MySQL, Redis)
- [x] Go module init for each service
- [x] Proto definitions
- [x] Generate Go code from proto

### Phase 2: Wiki Service (Basic CRUD) ✅
- [x] DB connection & migration
- [x] Repository layer
- [x] gRPC handler (Create, Read, Update, Delete)
- [x] Unit tests (model, handler with mock)

### Phase 3: Auth Service + JWT ✅
- [x] User registration (bcrypt)
- [x] Login & JWT generation (RS256)
- [x] RSA key pair generation
- [x] gRPC Reflection

### Phase 4: API Gateway ✅
- [x] REST → gRPC routing (Go 1.22 ServeMux)
- [x] Wiki REST endpoints
- [x] Auth REST endpoints

### Phase 5: CQRS Implementation (Wiki) ✅
- [x] Command/Query repository separation
- [x] Redis read model with fallback to MySQL
- [x] Cache TTL (10 minutes)

### Phase 6: Profile Service ✅
- [x] Proto definition & code generation
- [x] Profile model & repository
- [x] Portfolio item model & repository
- [x] gRPC handler
- [x] main.go with service registration
- [x] Gateway REST endpoints

### Phase 7: Frontend (Next.js) ✅
- [x] Project setup
- [x] Wiki page (article list + detail)
- [x] Admin page (article editor with React 19 action)
- [x] Profile page
- [x] Portfolio page (developing / completed state)
- [x] Dark mode (stone palette, system preference)
- [x] Keyboard shortcuts (? key + hover)
- [x] Gateway CORS middleware

### Phase 7.5: Frontend Enhancement
- [x] 戻るボタン（記事詳細ページにナビゲーション）
- [x] サイドバー追加（カテゴリ一覧 + スクロール対応）
- [x] 記事検索バー（タイトル/内容でフィルタ）
- [x] WikiページでFooter非表示
- [x] 記事一覧のキャッシュ問題修正（Next.js fetch cache: 'no-store'）

### Phase 7.6: Frontend Polish
- [x] Markdown 表示対応（react-markdown + @tailwindcss/typography）
- [x] Admin プレビュー付きエディタ（左右分割）
- [x] プレビュー拡大モーダル（中央に大きく表示）
- [x] Markdownヘルプパネル（DB設計書・API仕様・gRPC一覧・Mermaid図テンプレート + コピーボタン）
- [x] Mermaid図表対応（フローチャート・シーケンス図・ER図をMarkdown内で描画）
- [x] コードブロックのシンタックスハイライト（rehype-highlight）
- [x] 削除確認ダイアログ（ConfirmModal — 汎用コンポーネント）
- [x] 通知機能（Toast — 汎用コンポーネント、どこでも使える）
- [x] Changelog / Updates ページ

### Phase 7.7: カテゴリ階層（フル実装） ✅
- [x] DB: categories table 追加（id, name, parent_id）
- [x] DB: articles に category_id カラム追加
- [x] Proto: Category message + RPC（ListCategories, CreateCategory, DeleteCategory）
- [x] Proto: Article に category_id 追加
- [x] Backend: Category model / repository / handler（Go Wiki Service）
- [x] Backend: article INSERT/SELECT に category_id 追加
- [x] Gateway: Category REST endpoints（GET/POST/DELETE /api/categories）
- [x] Frontend: api.ts に createCategory / deleteCategory 追加
- [x] Frontend: CategoryManager コンポーネント（作成 + 一覧 + 削除 + ConfirmModal）
- [x] Frontend: Admin ページにタブ切り替え（記事作成 / カテゴリ管理）
- [x] Frontend: CategoryManager ルート/子の視覚的区別（アイコン + インデント）
- [x] Frontend: サイドバーにカテゴリツリー表示（展開/折りたたみ）
- [x] Frontend: サイドバー「すべて」ボタン + 選択中ハイライト
- [x] Frontend: 記事一覧をカテゴリで絞り込み（URL パラメータ ?category=id）
- [x] Frontend: 記事カードUI改善（stripMarkdown + ホバーアニメーション + 空状態UI）
- [x] Frontend: Changelog ページを Markdown ファイルベースに移行

### Phase 7.8: Profile ページリデザイン ✅
- [x] DB: profiles に avatar_url / twitter_url / linkedin_url / skills カラム追加
- [x] Proto: Profile message に新フィールド追加（avatar_url, twitter_url, linkedin_url, skills）
- [x] Proto: UpdateProfileRequest / CreateProfileRequest に新フィールド追加
- [x] Backend: Profile model / repository / handler 更新
- [x] Backend: Gateway handler 更新（新フィールドの JSON mapping）
- [x] Frontend: Profile interface 更新（types.ts）
- [x] Frontend: Profile ページリデザイン（Hero + Skills + About セクション）
- [x] Frontend: motion（旧 Framer Motion）でアニメーション追加
- [x] Frontend: Skills に react-icons/si アイコンマッピング
- [x] Frontend: ProfileManager に新フィールド入力欄追加
- [x] Frontend: Server Action 更新（Skills カンマ区切り → JSON 変換）

### Phase 7.9: Portfolio リデザイン & Wiki 目次 ✅
- [x] Portfolio ページリデザイン（Client Component + スライダー + カードレイアウト）
- [x] DB: portfolio_items に category / tech_stack カラム追加
- [x] Proto: PortfolioItem / Create/UpdateRequest に category / tech_stack 追加
- [x] Backend: PortfolioItem model / repository / handler / gateway 更新
- [x] Frontend: PortfolioCard にステータス + カテゴリー + Tech Stack タグ表示
- [x] Frontend: CardSlider（CSS snap + ナビボタン、4枚以上で表示）
- [x] Frontend: PortfolioManager CRUD + カテゴリー選択 + Tech Stack 入力
- [x] Frontend: Admin ページに「ポートフォリオ」タブ追加
- [x] Wiki 記事詳細ページに目次（TOC）サイドバー追加
- [x] ArticleContent の h2/h3 に slug ID 付与
- [x] TableOfContents（IntersectionObserver で現在地ハイライト + スムーズスクロール）

### Phase 7.10: トップページリデザイン & ブランディング ✅
- [x] Gateway: CreateProfile / UpdateProfile に wantedly_url 追加
- [x] Backend: Profile model / repository に WantedlyURL 追加
- [x] Frontend: Profile interface に wantedly_url 追加
- [x] Frontend: ProfileManager に Wantedly URL 入力欄追加
- [x] Frontend: Profile ページに Wantedly リンク表示（SiWantedly アイコン）
- [x] Frontend: Profile ページ About セクション削除（Bio 重複解消）
- [x] Frontend: トップページをハードコード Hero セクションにリデザイン
- [x] Frontend: フローティング技術キーワード背景（motion フェードイン）
- [x] Frontend: Values をマインドマップ風 UI に変更
- [x] Frontend: What is TenHub セクション + ホバー下線アニメーション
- [x] Frontend: ProtocolBuffers カスタム SVG アイコン追加
- [x] Frontend: Navbar にロゴ画像 + TenHub ブランディング
- [x] プロジェクト名を knowhub → TenHub に変更

### Phase 8: Polish + Deploy (VPS)
- [x] JWT 認証ミドルウェア（Gateway）
- [x] Login / Register で HttpOnly Cookie セット
- [x] slog structured logging + ログローテーション（lumberjack）
- [x] Graceful Shutdown（シグナルキャッチ + DB Close）
- [x] Error response cleanup (hide internals)
- [x] CORS 環境変数化
- [x] Swagger/OpenAPI ドキュメント生成（Gateway）
- [x] フロントエンド認証（Login / Logout / Auth Guard / Route Handlers）
- [x] Server Action → Client Component + api.ts 移行（Node.js Cookie 禁止ヘッダー問題の解消）
- [x] Gateway Auth middleware: Cookie + Authorization Bearer 両対応
- [x] Wiki 記事の公開/限定公開（visibility: public / locked）
- [ ] VPS setup (SSH, firewall)
- [ ] Nginx + Let's Encrypt
- [ ] Docker Compose production config
- [ ] GitHub Actions CI/CD

### Phase 9: AI Service
- [x] AI Service 雛形（Proto, Go module）
- [x] LLM プロバイダー抽象化インターフェース（provider.go）
- [x] 検索エンジン抽象化インターフェース（search.go）
- [x] AI Service main.go + gRPC handler
- [x] TF-IDF スクラッチ実装（トークン化 → TF計算 → IDF計算 → TF-IDFベクトル → コサイン類似度）
- [x] トークン化の日本語対応（文字種境界分割: 漢字/ひらがな/カタカナ/Latin/数字）
- [x] Ollama プロバイダー実装（ollama.go）
- [x] GLM-5 プロバイダー実装（glm5.go、OpenAI 互換フォーマット）
- [x] OpenAI プロバイダー実装（openai.go、オプション）
- [x] Gemini プロバイダー実装（gemini.go、OpenAI 互換フォーマット）
- [x] DeepSeek プロバイダー実装（deepseek.go、OpenAI 互換フォーマット）
- [x] BM25 スクラッチ実装（TF-IDF拡張、文書長正規化、k1/b パラメータ）
- [x] テストコード（TF-IDF / BM25 / tokenize / 各計算関数）
- [x] Gateway REST endpoints（POST /api/ai/search, /api/ai/summarize, /api/ai/ask）
- [x] Article search 動作確認（gRPC / REST テスト）
- [x] Article summarization（LLM による要約）
- [x] Q&A based on wiki content（RAG: 検索結果をコンテキストに LLM 回答）
- [x] Ollama モデル設定の環境変数化（OLLAMA_MODEL）
- [ ] Article search with vector embeddings（Ollama embedding モデル）
- [ ] Chat interface（フロントエンド）
- [ ] Gateway AI endpoints auth スキップ（認証なしでアクセス可能）
- [ ] RAG sources に記事タイトル付きリンク表示（Proto 変更: Source message）
- [ ] Chat で LLM モデル選択 UI（ユーザーがモデル切替）
- [ ] ユーザー API Key 入力（セキュアな取り扱い: マスク入力・サーバー非保存）
- [ ] Gateway AI timeout 設定（context canceled 対策）

### Phase 10: MCP Server
- [ ] MCP Server implementation (Go)
- [ ] Tools: create_article, search_articles, list_articles
- [ ] Resources: article content access
- [ ] Integration with Claude Desktop / other AI assistants
