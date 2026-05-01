# knowhub - Workflow & Progress

## Documentation Rules（ドキュメント更新ルール）

機能追加・修正・リファクタリングなどのコード変更を行った場合は、**必ず**以下のドキュメントを更新すること：

### 更新対象
| ファイル | タイミング | 更新内容 |
|----------|-----------|---------|
| `doc/backend/SPEC.md` | フェーズの完了・新フェーズの開始時 | 該当フェーズの `[ ]` → `[x]`、新規項目の追加 |
| `doc/backend/WORKFLOW.md` | 各タスクの完了時 | 該当フェーズの Progress に `[x]` を付ける、新しい Progress セクションの追加 |
| `frontend/src/lib/changelog.md` | **コード変更のたびに毎回** | 日付エントリを一番上に追記（新しい日付が上） |

### 更新フォーマット

**changelog.md** は以下の形式で **ファイルの先頭** に追記する：
```markdown
## YYYY-MM-DD 変更の概要タイトル
- 変更内容1
- 変更内容2
```

**SPEC.md** は該当フェーズのチェックボックスを `[x]` にする。新規項目があれば追加。

**WORKFLOW.md** は該当フェーズの Progress セクションに `[x]` を付ける。該当セクションがなければ新しく作る。

### いつ更新するか
- 新しい機能を実装したとき
- バグ修正を行ったとき
- リファクタリングを行ったとき
- UIコンポーネントを追加・変更したとき
- APIエンドポイントを追加・変更したとき
- DBスキーマを変更したとき

**変更完了後に必ず以下の3ファイルを更新すること：**
1. `doc/backend/SPEC.md` — 該当フェーズのチェックボックス更新・新規項目追加
2. `doc/backend/WORKFLOW.md` — 該当フェーズの Progress に `[x]` を付ける・新しい Progress セクション追加
3. `frontend/src/lib/changelog.md` — 日付エントリを一番上に追記（内容が既存エントリと重複しないように注意）

## Working Rules（AIとの協業ルール）

- [x] **AIの役割は「教師」** — 実装のヒントを詳細に提供し、実装力を高めるように手引きする
- [x] **ヒントにはコード例＋説明を必ず含める**（コード例なしだと理解できない）
- [x] **ユーザーが自分でコードを書く** — AIは書かない、ユーザーがアウトプットする
- [x] **AIがコードをレビューする** — 書けたらレビュー、修正の繰り返しで実装力を伸ばす
- [x] **最後にリファクタリング** — 各フェーズ完了時に、実務を意識した最適な実装にブラッシュアップする
- [x] このスタイルでSPEC.mdの各フェーズを完了させる

## Git Rules

- [x] コミットメッセージは Conventional Commits 形式（`feat:`, `fix:`, `refactor:` など）
- [x] AIはコミット内容（メッセージ）だけ提案し、実行はユーザーが判断する

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

### Frontend
- [x] Category interface は `@/lib/types` に統一（Sidebar.tsx のローカル定義を削除）
- [x] Changelog データは `changelog.md`（Markdown）で管理（DB不要・上に追記するだけ）
- [x] カテゴリ絞り込みは URL パラメータ方式（`/wiki?category=abc`）を採用

## Phase 6 Progress（Profile Service）

- [x] Proto definition & code generation
- [x] Profile model
- [x] PortfolioItem model
- [ ] PortfolioItem model の Status → PortfolioStatus カスタム型適用（TODO）
- [x] Profile repository
- [x] PortfolioItem repository
- [ ] gRPC handler
- [ ] main.go with service registration
- [ ] Migration のテーブル名修正（`articles` → `portfolio_items`）済み

## Phase 7 Progress（Frontend）

- [x] Project setup
- [x] Wiki page (article list + detail)
- [x] Admin page (article editor)
- [x] Profile page
- [x] Portfolio page (developing / completed state)
- [x] Dark mode

## Phase 7.5 Progress（Frontend Enhancement）

- [x] 戻るボタン（記事詳細ページにナビゲーション）
- [x] サイドバー追加（カテゴリ一覧 + スクロール対応）
- [x] 記事検索バー（タイトル/内容でフィルタ）
- [x] WikiページでFooter非表示
- [x] 記事一覧のキャッシュ問題修正（cache: 'no-store'）

## Phase 7.6 Progress（Frontend Polish）

- [x] Markdown 表示対応（react-markdown + @tailwindcss/typography）
- [x] Admin プレビュー付きエディタ（左右分割）
- [x] プレビュー拡大モーダル（中央に大きく表示）
- [x] Markdownヘルプパネル（DB設計書・API仕様・gRPC一覧・Mermaid図 + コピーボタン）
- [x] Mermaid図表対応（フローチャート・シーケンス図・ER図）
- [x] コードブロック シンタックスハイライト（rehype-highlight）
- [x] 削除確認ダイアログ（ConfirmModal — 汎用コンポーネント）
- [x] 通知機能（Toast — 汎用コンポーネント、どこでも使える）
- [x] Changelog / Updates ページ

## Phase 7.7 Progress（カテゴリ階層）

- [x] DB: categories table（id, name, parent_id）
- [x] DB: articles に category_id 追加、github_url 削除
- [x] Proto: Category message + RPC（ListCategories, CreateCategory, DeleteCategory）
- [x] Proto: Article に category_id 追加
- [x] Backend: Category model / repository / handler
- [x] Backend: article INSERT/SELECT に category_id 追加
- [x] Gateway: Category REST endpoints（GET/POST/DELETE /api/categories）
- [x] Frontend: api.ts に createCategory / deleteCategory 追加
- [x] Frontend: CategoryManager コンポーネント（作成 + 一覧 + 削除 + ConfirmModal）
- [x] Frontend: Admin ページにタブ切り替え（記事作成 / カテゴリ管理）
- [x] Frontend: CategoryManager ルート/子の視覚的区別（FiFolder / FiFile + インデント）
- [x] Frontend: Sidebar buildTree で API からツリー表示（展開/折りたたみ）
- [x] Frontend: Sidebar CategoryItem スタイル修正（... → 実際の Tailwind クラス）
- [x] Frontend: Sidebar 「すべて」ボタン + 選択中ハイライト
- [x] Frontend: WikiClient カテゴリ絞り込み（URL パラメータ ?category=id）
- [x] Frontend: WikiClient 記事カードUI改善（stripMarkdown + ホバーアニメーション + 空状態UI）
- [x] Frontend: Changelog ページを Markdown ファイルベースに移行（fs.readFileSync + parseChangelog）

## Phase 7.8 Progress（Profile ページリデザイン）

- [x] DB: profiles に avatar_url / twitter_url / linkedin_url / skills カラム追加
- [x] Proto: Profile / UpdateProfileRequest / CreateProfileRequest に新フィールド追加
- [x] Backend: Profile model に新フィールド追加（NewProfile / Update 更新）
- [x] Backend: Profile repository クエリ更新（SELECT / INSERT / UPDATE）
- [x] Backend: Profile handler 変換更新（toProfile）
- [x] Backend: Gateway handler JSON 項目追加
- [x] Frontend: types.ts Profile interface 更新
- [x] Frontend: api.ts updateProfile 引数更新
- [x] Frontend: Profile ページリデザイン（Hero + Skills + About + motion アニメーション）
- [x] Frontend: ProfileManager 新フィールド入力欄追加

## Phase 7.9 Progress（Portfolio リデザイン & Wiki 目次）

- [x] Portfolio ページを Client Component 化 + motion アニメーション
- [x] DB: portfolio_items に category / tech_stack カラム追加
- [x] Proto: PortfolioItem / Create/UpdateRequest に category / tech_stack 追加
- [x] Backend: PortfolioItem model / repository / handler / gateway 更新
- [x] Frontend: PortfolioCard（ステータス + カテゴリー + Tech Stack タグ）
- [x] Frontend: CardSlider（CSS snap + ナビボタン）
- [x] Frontend: PortfolioManager CRUD + Admin タブ追加
- [x] Wiki 記事詳細に目次（TOC）サイドバー追加
- [x] ArticleContent h2/h3 に slug ID 付与
- [x] TableOfContents（IntersectionObserver + スムーズスクロール）
- [x] Frontend: Server Action 更新（Skills カンマ区切り → JSON 変換）

## Phase 7.10 Progress（トップページリデザイン & Wantedly & ブランディング）

- [x] Gateway handler: CreateProfile / UpdateProfile に WantedlyUrl 追加
- [x] Backend: Profile model NewProfile / Update に WantedlyURL 追加
- [x] Backend: Profile repository INSERT / UPDATE に wantedly_url 追加
- [x] Frontend: types.ts Profile interface に wantedly_url 追加
- [x] Frontend: api.ts updateProfile 引数に wantedly_url 追加
- [x] Frontend: Server Action profile.ts に wantedly_url 送信追加
- [x] Frontend: ProfileManager に Wantedly URL 入力欄追加
- [x] Frontend: Profile ページに Wantedly リンク追加（SiWantedly アイコン）
- [x] Frontend: Profile ページ About セクション削除（Bio 重複解消）
- [x] Frontend: トップページをプロフィール依存からハードコード Hero にリデザイン
- [x] Frontend: Hero セクション（knowhub タイトル + グラデーション + タグライン + タイピングライン）
- [x] Frontend: フローティング技術キーワード背景（motion フェードイン）
- [x] Frontend: Values をマインドマップ風にリデザイン（中心 Mindset ノード + 放射状 pill ノード）
- [x] Frontend: What is knowhub? セクション（ホバーで左→右 下線アニメーション）
- [x] Frontend: Skills に ProtocolBuffers カスタム SVG アイコン追加
- [x] Frontend: Navbar にロゴ画像追加 + TenHub ブランディング
- [x] プロジェクト名を knowhub → TenHub に変更

## Phase 8.1 Progress（Wiki 記事 visibility: 公開/限定公開）

- [x] DB: articles に visibility カラム追加（VARCHAR(20) DEFAULT 'public'）
- [x] Proto: Article / CreateArticleRequest / UpdateArticleRequest に visibility 追加
- [x] Backend: Article model に Visibility フィールド追加（NewArticle / Update 更新）
- [x] Backend: article_command.go INSERT/UPDATE に visibility 追加
- [x] Backend: article_query.go SELECT / Scan に visibility 追加
- [x] Backend: article.go（旧リポ）同パターンで更新
- [x] Backend: wiki_cqrs.go / wiki.go Create/Update/toProto に visibility 追加
- [x] Backend: Gateway wiki_handler.go Create/Update JSON に visibility 追加
- [x] Backend: swagger/types.go に visibility 追加
- [x] Frontend: types.ts Article interface に visibility 追加
- [x] Frontend: api.ts saveArticle に visibility 引数追加
- [x] Frontend: Editor.tsx に公開設定セレクトボックス追加（一般公開 / 限定公開）
- [x] Frontend: WikiClient.tsx に鍵アイコンバッジ + プレビューマスク表示
- [x] Frontend: wiki/[id]/page.tsx に鍵アイコン + コンテンツぼかしオーバーレイ + TOC非表示
- [x] Test: article_test.go に visibility テスト追加（DefaultVisibility / LockedVisibility / Update_Visibility）

## Backend TODO（CQRS キャッシュ無効化）

- [x] article_command.go に rdb 追加（Create/Save/Delete後にRedisキャッシュ削除）
- [x] main.go の初期化で rdb を commandRepo に渡す
- [x] article_query.go の FindAll に ORDER BY created_at DESC 追加済み

## Phase 8 Progress（Polish + Deploy）

- [x] JWT Cookie / CORS / バリデーション（Gateway）
  - [x] Auth Service: VerifyToken ハンドラー実装
  - [x] Auth Repository: FindByID 追加
  - [x] Gateway: JWT 認証ミドルウェア（Cookie検証 + context に userID 保存）
  - [x] Gateway: CORS → Auth → Router ミドルウェアチェーン構築
- [x] Login / Register で HttpOnly Cookie セット
- [x] slog structured logging
  - [x] 共通パッケージ pkg/logger 作成（slog + lumberjack ローテーション）
  - [x] 共通パッケージ pkg/server 作成（Graceful Shutdown + DB Close）
  - [x] 全サービス main.go を slog + Graceful Shutdown 対応
  - [x] config に LogLevel 追加（環境変数 LOG_LEVEL）
  - [x] DB Ping ヘルスチェック追加
  - [x] Gateway: CORS を middleware パッケージに分離 + dialService 関数共通化
- [x] Error response cleanup (hide internals)
  - [x] Gateway handler: 内部エラーを slog.Error でログ記録、クライアントには汎用メッセージ（"internal server error"等）を返す
  - [x] Auth middleware: トークン検証失敗時も内部詳細を隠蔽
- [x] CORS 環境変数化
  - [x] Gateway: config パッケージ作成（CORS / サービスアドレス / ポート / LogLevel）
  - [x] CoreMiddleware: AllowedOrigin / AllowedMethods / AllowedHeaders を構造体フィールド化
  - [x] main.go: config.Load → Logger → gRPC → CORS チェーン構築
  - [x] .env: ALLOWED_ORIGIN / ALLOWED_METHODS / ALLOWED_HEADERS / AUTH_ADDR / WIKI_ADDR / PROFILE_ADDR / GATEWAY_PORT 追加
- [x] Swagger/OpenAPI ドキュメント生成（Gateway）
  - [x] swag + http-swagger パッケージ導入
  - [x] main.go に @title / @version / @host アノテーション追加 + /swagger/ ルート追加
  - [x] auth / wiki / profile ハンドラーに全18個のアノテーション追加
  - [x] swagger/types.go: リクエスト構造体（9個）を別パッケージに定義 + example 値
  - [x] @Param をインライン object{} → swagger.TypeName 構造体参照に変更
  - [x] Auth middleware: /swagger/ プレフィックスを認証ホワイトリストに追加
- [x] フロントエンド認証（Login / Logout / Auth Guard / Route Handlers）
  - [x] Gateway: Auth middleware GET スキップ（GET /api/user/me は認証あり）
  - [x] Gateway: Login/Register body に token 追加（Route Handler 用）
  - [x] Gateway: FindByID RPC 追加（Proto / Handler / Repository）
  - [x] Gateway: /api/user/me エンドポイント追加
  - [x] Frontend: Next.js rewrites で /api/* → Gateway プロキシ（Cookie 自動転送）
  - [x] Frontend: Route Handlers（/api/auth/login / register / logout）で Cookie セット/削除
  - [x] Frontend: AuthContext（checkAuth / login / logout）+ AuthProvider
  - [x] Frontend: Login ページ作成（Route Handler 経由でログイン）
  - [x] Frontend: Admin ページ認証ガード（isLoggedIn チェック → 未ログインは /login へリダイレクト）
  - [x] Frontend: Navbar 条件分岐（ログイン時のみ Admin / Logout 表示）
  - [x] Frontend: api.ts の API_BASE を '/api' に統一（credentials 不要）
  - [x] Frontend: Server Actions に fetchWithAuth ヘルパー（cookies() で手動 Cookie 転送）
- [x] Server Action → Client Component + api.ts 移行
  - [x] Node.js fetch の相対URL問題修正（Server: 直URL / Client: 相対URL 自動切替）
  - [x] Node.js fetch の Cookie 禁止ヘッダー問題解消（Server Action → Client Component に移行）
  - [x] api.ts に mutation 関数統合（saveArticle / deleteArticle / saveProfile / savePortfolioItem / deletePortfolioItem）
  - [x] Editor.tsx: useActionState → onSubmit + api.ts
  - [x] ProfileManager.tsx: useActionState → onSubmit + api.ts
  - [x] PortfolioManager.tsx: ハードコード URL → api.ts
  - [x] Server Action ファイル（article.ts / profile.ts）をアーカイブ化
  - [x] Gateway: Auth middleware を Cookie + Authorization Bearer 両対応に変更
  - [x] .env タイプミス修正（PROFILE_ADDR / ALLOWED_CREDENTIALS / LOG_LEVEL）
- [ ] VPS setup (SSH, firewall)
- [ ] Nginx + Let's Encrypt
- [ ] Docker Compose production config
- [ ] GitHub Actions CI/CD

## Phase 9 Progress（AI Service）

- [x] Proto: ai.proto 定義（SearchArticles / SummarizeArticle / AskQuestion）
- [x] Proto: Go コード生成（ai.pb.go / ai_grpc.pb.go）
- [x] LLM プロバイダー抽象化インターフェース（provider.go: Generate / Chat）
- [x] Ollama プロバイダー実装（ollama.go: /api/generate + /api/chat）
- [x] GLM-5 プロバイダー実装（glm5.go: OpenAI 互換フォーマット）
- [x] OpenAI プロバイダー実装（openai.go: オプション）
- [x] Gemini プロバイダー実装（gemini.go: OpenAI 互換フォーマット）
- [x] DeepSeek プロバイダー実装（deepseek.go: OpenAI 互換フォーマット）
- [x] 検索エンジン抽象化インターフェース（search.go: Document / SearchResult / SearchEngine）
- [x] TF-IDF スクラッチ実装（tfidf.go）
  - [x] tokenize: 小文字化 → 文字種境界分割（漢字/ひらがな/カタカナ/Latin/数字）
  - [x] computeTF: 出現回数 / 全単語数
  - [x] computeIDF: log((N+1)/(df+1))、seen で重複カウント防止
  - [x] buildVocabulary: 単語 → インデックスの map 構築
  - [x] buildTFIDFVector: TF × IDF を vocabulary インデックスに配置
  - [x] cosineSimilarity: 内積 / (L2ノルムA × L2ノルムB)
  - [x] Index: トークン化 → 語彙構築 → IDF → TF-IDF ベクトル
  - [x] Search: クエリベクトル化 → コサイン類似度 → スコア降順ソート
- [x] AI Service config（config.go: LLM_PROVIDER / OLLAMA_URL / GLM5_API_KEY 等）
- [x] AI Service handler（ai.go: SearchArticles / SummarizeArticle / AskQuestion）
- [x] AI Service main.go（Wiki Service gRPC client + DI + Graceful Shutdown）
- [x] BM25 スクラッチ実装（bm25.go）
  - [x] BM25Engine 構造体（k1 / b パラメータ）
  - [x] computeTermFreq: 生の出現回数カウント
  - [x] computeBM25IDF: log((N - n + 0.5) / (n + 0.5) + 1)
  - [x] Index: トークン化 → docLens → avgDl → 語彙構築 → BM25 IDF
  - [x] Search: クエリ各単語で BM25 スコア計算 → 合算 → ソート
- [x] テストコード（tfidf_test.go / bm25_test.go）
- [x] Gateway REST endpoints（ai_handler.go + main.go ルート追加）
  - [x] POST /api/ai/search（SearchArticles）
  - [x] POST /api/ai/summarize（SummarizeArticle）
  - [x] POST /api/ai/ask（AskQuestion）
- [x] Gateway config に AIAddr 追加
- [x] .env に AI_ADDR / LLM_PROVIDER / OLLAMA_URL 等追加
- [x] Article search 動作確認（grpcurl で BM25 検索確認）
- [x] Article summarization 動作確認（Ollama + qwen3:1.7b で要約確認）
- [x] Q&A based on wiki content（RAG: 検索→コンテキスト→LLM回答、sources 返却確認）
- [x] Ollama モデル設定の環境変数化（OLLAMA_MODEL）
- [x] RAG sources に記事タイトル付きリンク（Proto: Source message 追加 + handler 更新）
- [x] Chat interface（フロントエンド）
  - [x] ChatInterface コンポーネント（ReactMarkdown + アイコン + スクロール制御）
  - [x] api.ts askQuestion 追加（model / apiKey パラメータ + AskSource 型定義）
  - [x] Chat ページルーティング（/chat）
  - [x] ダークモード prose スタイル調整（コードブロック・リスト・区切り線）
  - [x] チャット履歴永続化（localStorage + useSyncExternalStore + React 19 対応）
  - [x] 履歴削除機能（FiTrash2 + ConfirmModal）
  - [x] LLM モデル選択 UI（セレクトボックス + const.ts MODELS 定義）
  - [x] API Key 入力（password + sessionStorage、タブ閉じで消去）
- [x] Proto: QuestionRequest に model / api_key フィールド追加
- [x] Backend: LLM 動的プロバイダー選択（NewProvider ファクトリ + model prefix 判定）
- [x] Backend: AI handler でリクエストの model/api_key からプロバイダーを動的生成
- [x] Gateway: AI endpoints 認証スキップ（/api/ai/ プレフィックスで認証バイパス）
- [x] Gateway: AI ask endpoint に 60s timeout（context.WithTimeout）
- [x] キーボードショートカットから Admin（a キー）を削除

## Phase 9.5 Progress（Advanced Search: Vector + Graph RAG）

### Vector Embeddings
- [ ] Ollama embedding API クライアント（/api/embed エンドポイント）
  - [ ] EmbeddingProvider インターフェース定義（GetEmbedding / GetEmbeddings）
  - [ ] OllamaEmbeddingProvider 実装（http.Post → []float64）
  - [ ] 外部 API フォールバック（DeepSeek / OpenAI / Gemini embedding API）
- [ ] VectorEngine 構造体（SearchEngine インターフェース実装）
  - [ ] documents フィールド（元ドキュメント保持）
  - [ ] embeddings フィールド（[][]float64、インデックス時に一括生成）
  - [ ] Index: 全ドキュメント → embedding API → キャッシュ
  - [ ] Search: クエリ → embedding → コサイン類似度 → スコア降順
- [ ] cosineSimilarity（tfidf.go の流用 or 共通化）
- [ ] Hybrid Search（BM25 + Vector の重み付き統合）
  - [ ] HybridEngine 構造体（bm25 + vector エンジンを内包）
  - [ ] α * BM25正規化スコア + (1-α) * Vector スコア
  - [ ] min-max 正規化でスケールを統一
- [ ] Config に EmbeddingProvider / EmbeddingModel 追加
- [ ] main.go に "vector" / "hybrid" エンジン選択肢追加
- [ ] テストコード（embedding / cosine / vector search / hybrid）

### Graph RAG
- [ ] ナレッジグラフのデータ構造設計
  - [ ] Entity 構造体（ID / Name / Type / ArticleIDs）
  - [ ] Relation 構造体（Source / Target / Label）
  - [ ] KnowledgeGraph 構造体（entities map + relations slice + adjacency list）
- [ ] LLM によるエンティティ・リレーション抽出
  - [ ] プロンプト設計（記事 → JSON: {entities, relations}）
  - [ ] ExtractEntities 関数（LLM 呼び出し → JSON パース）
- [ ] GraphEngine 構造体（SearchEngine インターフェース実装）
  - [ ] Index: 記事 → エンティティ抽出 → グラフ構築
  - [ ] Search: クエリ → エンティティ特定 → BFS で関連ノード探索 → 関連記事収集
- [ ] Graph + Vector ハイブリッド回答生成
  - [ ] グラフ検索で関連記事を広く収集
  - [ ] Vector でセマンティック検索
  - [ ] 両方の結果をマージしてコンテキスト生成
- [ ] テストコード（グラフ構築 / BFS トラバーサル / 検索）

### フロントエンド: 検索エンジン選択 UI
- [ ] const.ts に SEARCH_ENGINES 定数追加
  - [ ] bm25: { id: 'bm25', name: 'BM25（キーワード検索）', needsKey: false }
  - [ ] vector: { id: 'vector', name: 'Vector（セマンティック検索）', needsKey: true }
  - [ ] hybrid: { id: 'hybrid', name: 'Hybrid（BM25 + Vector）', needsKey: true }
  - [ ] graph: { id: 'graph', name: 'Graph RAG（ナレッジグラフ）', needsKey: true }
- [ ] ChatInterface に検索エンジンセレクトボックス追加（LLM モデル選択の下）
- [ ] 選択したエンジンに応じて API Key 入力欄の表示/非表示切替
- [ ] api.ts askQuestion に search_engine / embedding_api_key パラメータ追加
- [ ] Proto: QuestionRequest に search_engine フィールド追加
- [ ] Backend: AI handler で search_engine 値から動的に SearchEngine を選択
  - [ ] "bm25" → BM25Engine（API Key 不要）
  - [ ] "vector" → VectorEngine（embedding API Key 使用）
  - [ ] "hybrid" → HybridEngine（BM25 + Vector）
  - [ ] "graph" → GraphEngine（LLM API Key 使用）

## Phase 10 Progress（MCP Server）

- [ ] MCP Server implementation (Go)
- [ ] Tools: create_article, search_articles, list_articles
- [ ] Resources: article content access
- [ ] Integration with Claude Desktop / other AI assistants
