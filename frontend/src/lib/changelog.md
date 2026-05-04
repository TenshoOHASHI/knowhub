## 2026-05-04 Agent モード改善 + UI ブラッシュアップ + Embedding 修正
- Backend: Agent 実行モード自動切替（外部モデル → 自律ReAct、Ollama → 固定パイプライン）
- Backend: isExternalModel ヘルパー追加（llm.NewProvider と同じプレフィックス判定ロジック）
- Backend: Embedding NewProvider ファクトリを model ベースルーティングに変更（DeepSeek APIキーの OpenAI 誤送信修正）
- Backend: Gateway タイムアウト調整（RAG: 60s→120s、Agent: 180s→300s）
- Frontend: チャット画面にシンタックスハイライト追加（rehype-highlight + github-dark.css）
- Frontend: チャット画面コードブロックにコピーボタン追加（CodeBlock コンポーネント）
- Frontend: ヘルプパネル追加（?ボタン + RAG/Agent モード別説明 + モード切替で自動閉じる）
- Frontend: Wiki 記事コードブロック背景を GitHub dark スタイルに変更（#161b22）
- Frontend: Wiki 記事コードブロックにコピーボタン追加（ArticleContent + Markdown 共通）
- Frontend: highlight.js/styles/github.css インポート削除（github-dark.css のみ使用）

## 2026-05-03 ReAct Agent + SearXNG 外部検索
- Proto: ai.proto に AskWithAgent RPC + AgentQuestionRequest / AgentStep / AgentSource / AgentQuestionResponse 追加
- Docker: docker-compose.yml に SearXNG サービス追加（:8888）
- Backend: config.go に SearXNGURL フィールド追加
- Backend: agent パッケージ新規作成（Tool interface + 5ツール: search_wiki/read_article/list_articles/web_search/read_url）
- Backend: ReAct ループ実装（max 10 iteration / Thought-Action-Observation パーサー / Final Answer 抽出）
- Backend: callbacks.go（OnToolStart/OnToolEnd/OnLLMStart/OnLLMEnd + slog ロギング）
- Backend: ai.go に AskWithAgent メソッド追加（searxngURL + ツールリスト構築 + agent.Run）
- Gateway: POST /api/ai/agent エンドポイント追加（timeout 180s）
- Frontend: api.ts に askWithAgent + AgentStep / AgentSource 型追加
- Frontend: const.ts に CHAT_MODES 定数追加（RAG / Agent）
- Frontend: AgentSteps.tsx 折りたたみ思考プロセス表示コンポーネント
- Frontend: ChatInterface.tsx モード切替セレクトボックス + Web検索チェックボックス + agentSteps 表示

## 2026-05-02 Markdown拡張: コールアウト + 折りたたみブロック
- Frontend: remark-callout プラグイン実装（Zenn記法 / GitHub記法 → コールアウトdiv変換）
- Frontend: Callout コンポーネント実装（7タイプ: note/info/tip/warning/caution/important/warm + SVGアイコン）
- Frontend: rehype-raw 追加（Markdown内HTMLタグレンダリング対応）
- Frontend: details/summary カスタムレンダラー（Tailwind スタイリング + ダークモード対応）
- Frontend: stripMarkdown 拡張（HTMLタグ/Zenn記法/コールアウトマーカー/setext下線除去）
- Frontend: MarkdownHelp にコールアウト・折りたたみブロック記法例追加
- Frontend: 共通 Markdown コンポーネント（Markdown.tsx）に Callout/div/details/summary 追加
- Frontend: EditorPreview に remarkCallout + preprocessCallouts + rehype-raw 統合

## 2026-05-02 Graph RAG 実装 + 検索エンジン動的選択リファクタリング
- Backend: graph.go 実装（Entity / Relation / KnowledgeGraph / GraphEngine / BFS 2-hop トラバーサル）
- Backend: LLM によるエンティティ・リレーション抽出（extractEntities / extractJSON / プロンプト設計）
- Backend: searchByTokens フォールバック（LLM 失敗時のトークン部分一致検索）
- Backend: embedding.NewProvider ファクトリ追加（apiKey から自動判定: 空→Ollama, sk-→OpenAI, AIza→Gemini, その他→GLM-5）
- Backend: search.SelectEngine ファクトリ追加（engineName → SearchEngine 自動生成）
- Proto: QuestionRequest に search_engine フィールド追加（field 4）
- Gateway: ai_handler.go AskQuestion に search_engine パラメータ追加
- Backend: main.go シンプル化（LLM/Embedding/Search Engine の switch 全削除、デフォルト Ollama + BM25 のみ）
- Backend: config.go 削減（SearchEngin / EmbeddingProvider / LLM 個別 API Key フィールド削除）
- Test: grpcurl で Graph RAG 動作確認（Ollama + graph で関連記事3件 + BM25 にない記事を発見）

## 2026-05-01 Hybrid Search 実装（BM25 + Vector 統合）
- Backend: HybridEngine 実装（SearchEngine インターフェース、BM25 + Vector を内包）
- Backend: normalizeScores（min-max 正規化でスコアを 0〜1 に統一）
- Backend: map[string]*hybridScore で記事ID マージ + α 重み付き統合スコア計算
- Backend: main.go に "hybrid" エンジン選択肢追加（α=0.5）
- Test: Hybrid 検索動作確認（BM25 の単語一致 + Vector の意味検索が統合され正確な順位を確認）

## 2026-05-01 Vector Search 実装（Embedding + VectorEngine + 多プロバイダー対応）
- Backend: EmbeddingProvider インターフェース定義（embedding/provider.go: GetEmbedding / GetEmbeddings）
- Backend: OllamaEmbeddingProvider 実装（/api/embed → []float64）
- Backend: OpenAI 互換共通 EmbeddingProvider 実装（OpenAI / DeepSeek / Gemini / GLM-5 対応）
- Backend: VectorEngine 実装（SearchEngine インターフェース、コサイン類似度検索）
- Backend: Config に EMBEDDING_PROVIDER / EMBEDDING_MODEL 追加
- Backend: main.go に "vector" エンジン選択肢追加（EmbeddingProvider → VectorEngine DI）
- Backend: UTF-8 安全な snippet 切り詰め修正（rune ベース、BM25/Vector 両対応）
- Backend: snippet の UTF-8 サニタイズ追加（strings.ToValidUTF8）
- Test: grpcurl で Vector 検索動作確認（gRPC / ユーザー認証）
- Test: GLM embedding-3 で日本語セマンティック検索精度が向上することを確認

## 2026-04-30 AI Chat 機能拡張（モデル選択・API Key・履歴永続化・認証スキップ）
- Frontend: ChatInterface に LLM モデル選択セレクトボックス追加（const.ts MODELS 定義）
- Frontend: API Key 入力欄追加（password + sessionStorage、タブ閉じで消去）
- Frontend: チャット履歴の localStorage 永続化（useSyncExternalStore + React 19 対応）
- Frontend: 履歴削除ボタン追加（FiTrash2 + ConfirmModal）
- Frontend: Chat ページタイトルを "Chat Bot" に変更
- Frontend: キーボードショートカットから Admin（a キー）を削除
- Frontend: ConfirmModal の quote スタイル統一
- Frontend: api.ts askQuestion に model / apiKey パラメータ追加、AskSource.articleId → article_id に変更
- Proto: QuestionRequest に model / api_key フィールド追加
- Backend: LLM 動的プロバイダー選択（NewProvider ファクトリ関数、model prefix で判定）
- Backend: AI handler でリクエストの model/api_key からプロバイダーを動的生成
- Gateway: AI endpoints の認証スキップ（/api/ai/ プレフィックス）
- Gateway: AI ask endpoint に 60s timeout 設定（context.WithTimeout）

## 2026-04-29 AI Chat Interface + LLM Provider 拡張
- Backend: Gemini プロバイダー実装（gemini.go、OpenAI 互換フォーマット）
- Backend: DeepSeek プロバイダー実装（deepseek.go、OpenAI 互換フォーマット）
- Backend: AI Service config に Gemini / DeepSeek 設定追加
- Backend: RAG Q&A のコンテキストを検索スニペット → 記事全文に改善
- Frontend: ChatInterface コンポーネント実装（ReactMarkdown + sources リンク付き）
- Frontend: api.ts に askQuestion 追加
- Frontend: Chat ページルーティング（/chat）
- Doc: Tailwind CSS パターン集追加（doc/frontend/tailwind-patterns.md）

## 2026-04-28 AI Service（TF-IDF / BM25 検索 + RAG Q&A）
- Backend: AI Service 雛形作成（Proto定義、Go module、gRPC handler）
- Backend: LLM プロバイダー抽象化インターフェース（provider.go: Generate / Chat）
- Backend: Ollama プロバイダー実装（ollama.go）
- Backend: GLM-5 プロバイダー実装（glm5.go、OpenAI 互換フォーマット）
- Backend: OpenAI プロバイダー実装（openai.go、オプション）
- Backend: TF-IDF スクラッチ実装（トークン化 → TF → IDF → コサイン類似度）
- Backend: BM25 スクラッチ実装（TF-IDF拡張、k1/b パラメータ）
- Backend: テストコード（TF-IDF / BM25 / tokenize / 各計算関数）
- Backend: Gateway REST endpoints（POST /api/ai/search, /api/ai/summarize, /api/ai/ask）
- Backend: Article search / summarization / RAG Q&A 動作確認

## 2026-04-27 Wiki 記事の公開/限定公開（visibility）機能
- DB: articles に visibility カラム追加（VARCHAR(20) DEFAULT 'public'）
- Proto: Article / Create/UpdateRequest に visibility フィールド追加 + Go コード再生成
- Backend: Article model（NewArticle / Update）に Visibility 追加、無効値は "public" にフォールバック
- Backend: CQRS / 旧リポの INSERT/UPDATE/SELECT クエリに visibility 追加
- Backend: gRPC Handler（wiki_cqrs / wiki）の Create/Update/toProto に visibility マッピング追加
- Gateway: wiki_handler.go の Create/Update JSON body に visibility 追加、swagger/types.go 更新
- Frontend: Article interface / saveArticle 引数に visibility 追加
- Frontend: Editor に公開設定セレクトボックス追加（一般公開 / 限定公開）
- Frontend: Wiki 一覧で locked 記事に鍵アイコンバッジ + プレビュー文マスク表示
- Frontend: 記事詳細で locked 時にぼかしオーバーレイ + TOC 非表示
- Test: visibility のデフォルト値・locked・Update に関するテストケース追加

## 2026-04-27 Server Action → Client Component + api.ts 移行
- Frontend: api.ts に Server/Client 自動切替追加（typeof window で絶対URL/相対URL切替）
- Frontend: mutation 関数を api.ts に統合（saveArticle / deleteArticle / saveProfile / savePortfolioItem / deletePortfolioItem）
- Frontend: Editor.tsx を useActionState → onSubmit + api.ts に変更
- Frontend: ProfileManager.tsx を useActionState → onSubmit + api.ts に変更
- Frontend: PortfolioManager.tsx のハードコード URL → api.ts に変更
- Frontend: Server Action ファイル（article.ts / profile.ts）をアーカイブ化
- Gateway: Auth middleware を Cookie + Authorization Bearer 両対応に変更
- .env: タイプミス修正（PROFILE_ADDR / ALLOWED_CREDENTIALS / LOG_LEVEL 二重定義）

## 2026-04-26 フロントエンド認証（Login / Logout / Auth Guard）
- Gateway: Auth middleware GET スキップ追加（GET /api/user/me は認証必要）
- Gateway: Login/Register レスポンス body に token 追加
- Gateway: FindByID RPC 追加（Proto 再生成 + Auth Handler + Repository）
- Gateway: /api/user/me エンドポイント追加
- Frontend: Next.js rewrites で /api/* → Gateway プロキシ（Cookie 自動転送）
- Frontend: Route Handlers（login / register / logout）で HttpOnly Cookie セット/削除
- Frontend: AuthContext（checkAuth / login / logout）+ AuthProvider（layout.tsx）
- Frontend: Login ページ作成
- Frontend: Admin ページ認証ガード（isLoggedIn チェック → /login リダイレクト）
- Frontend: Navbar ログイン状態でリンク切り替え（Admin / Logout はログイン時のみ）
- Frontend: api.ts の API_BASE を '/api' に統一
- Frontend: Server Actions に fetchWithAuth 追加（cookies() で Gateway に Cookie 転送）

## 2026-04-26 Swagger/OpenAPI ドキュメント生成
- Gateway: swag + http-swagger パッケージ導入
- Gateway: main.go に Swagger アノテーション追加 + /swagger/ ルート追加
- Gateway: auth / wiki / profile ハンドラーに全18個の API アノテーション追加
- Gateway: swagger/types.go にリクエスト構造体（9個）を別パッケージで定義
- Gateway: @Param をインライン object{} → swagger.TypeName 構造体参照に変更
- Gateway: Auth middleware に /swagger/ プレフィックスを認証ホワイトリストに追加

## 2026-04-26 CORS 環境変数化 & Error response cleanup
- Gateway: config パッケージ作成（CORS / サービスアドレス / ポート / LogLevel を .env で管理）
- Gateway: CoreMiddleware の CORS 設定を環境変数から読み込むように変更
- Gateway: main.go の Logger 初期化順序を config.Load → loggerpkg.New に修正
- Gateway: ポートを cfg.Port に統一（ハードコード解消）
- 全サービス: dbutil.Wrap(db) で DB クエリログ出力（wiki / profile / auth）

## 2026-04-25 slog構造化ログ & Graceful Shutdown
- 共通パッケージ `pkg/logger` 作成（slog + lumberjack ログローテーション）
- 共通パッケージ `pkg/server` 作成（Graceful Shutdown + DB Close）
- 全サービス（Auth / Wiki / Profile / Gateway）の main.go を slog + Graceful Shutdown 対応
- 全サービスの config に LogLevel 追加（環境変数 LOG_LEVEL で切替）
- DB Ping ヘルスチェック追加（起動時に接続確認）
- Gateway: CORS ミドルウェアを middleware パッケージに分離（CoreMiddleware）
- Gateway: dialService 関数で gRPC 接続を共通化 + 接続状態ログ出力
- `log.Printf` / `log.Fatalf` / `panic` を `slog.Info` / `slog.Error` + `os.Exit(1)` に統一

## 2026-04-25 JWT認証ミドルウェア & UI修正
- Auth Service に VerifyToken ハンドラー実装（gRPC）
- Auth Repository に FindByID 追加
- Gateway に JWT 認証ミドルウェア追加（Cookie検証 + context に userID 保存）
- Gateway main.go に CORS → Auth → Router のミドルウェアチェーン構築
- Portfolio ページに作成日表示追加（created_at）
- PortfolioCard にカレンダーアイコン付き日付表示追加
- PortfolioManager 一覧をページネーション対応（3件/ページ + ドットインジケーター）
- トップページ Hub グラデーションをモノクロームに変更（from-stone-700 to-stone-400）

## 2026-04-25 トップページリデザイン & ブランディング
- プロジェクト名を knowhub → TenHub に変更
- トップページをハードコード Hero セクションにリデザイン（プロフィール依存を削除）
- Hero セクション（グラデーションタイトル + タグライン + タイピングライン）
- フローティング技術キーワード背景（Go, gRPC, CQRS, Next.js など）
- Values をマインドマップ風 UI に変更（中心 Mindset ノード + 放射状 pill ノード）
- What is TenHub セクション（ホバーで左→右 下線アニメーション）
- ProtocolBuffers カスタム SVG アイコン追加（react-icons/si に無いため）
- Navbar にロゴ画像追加 + TenHub ブランディング
- Profile に wantedly_url 追加（全層: Proto / Model / Repository / Handler / Gateway / Frontend）
- ProfileManager に Wantedly URL 入力欄追加
- Profile ページ About セクション削除（Bio 重複解消）

## 2026-04-24 Portfolio リデザイン & Wiki 目次追加
- Portfolio ページを Client Component 化（スライダー + カードレイアウト）
- Portfolio に category / tech_stack カラム追加（DB migration）
- PortfolioCard にステータスバッジ + カテゴリーバッジ + Tech Stack タグ追加
- CardSlider コンポーネント実装（CSS snap + ナビボタン、4枚以上で表示）
- PortfolioManager にカテゴリー選択・Tech Stack 入力欄追加
- Admin ページに「ポートフォリオ」タブ追加（CRUD対応）
- Wiki 記事詳細ページに目次（TOC）サイドバー追加
- ArticleContent の h2/h3 に ID 付与（slugify）
- TableOfContents コンポーネント（IntersectionObserver で現在地ハイライト）

## 2026-04-24 Profile ページリデザイン
- Profile に avatar_url / twitter_url / linkedin_url / skills カラム追加（DB migration）
- Proto フィールド追加（Profile, UpdateProfileRequest, CreateProfileRequest）
- バックエンド model / repository / handler / gateway 更新
- Profile ページを Client Component 化（Hero / Skills / About セクション）
- motion ライブラリでフェードイン・スタガー・スクロールアニメーション追加
- Skills に react-icons/si アイコンマッピング追加
- ProfileManager に新フィールド入力欄追加（Avatar / Twitter / LinkedIn / Skills）
- Server Action 更新（新フィールド送信 + Skills JSON 変換）

## 2026-04-24 コンポーネント分割 & カテゴリ連携
- Admin ページをコンポーネント分割（Editor, EditorPreview, Markdown, CategoryManager に分離）
- Server Action 追加（`actions/article.ts`）で記事更新ロジックを分離
- Sidebar を API 連携のカテゴリツリーに刷新（ハードコード削除 → `getCategories`）
- WikiClient にカテゴリ絞り込み追加（URL パラメータ `?category=id`）
- WikiClient 記事カードUI改善（stripMarkdown + ホバーアニメーション + 空状態UI）
- Admin タブ切り替え（記事作成 / カテゴリ管理）

## 2026-04-23 カテゴリ管理 & UI改善 & CRUD改良
- カテゴリ作成・削除機能（Admin タブ追加）
- 削除確認モーダル（ConfirmModal 統合）
- ルート/子カテゴリの視覚的区別（アイコン + インデント）
- サイドバー カテゴリツリー表示（API連携 + buildTree）
- Wikiページの更新機能追加（Server Action）

## 2026-04-22 Markdown対応 & エディタ強化
- Markdown → HTML レンダリング（react-markdown）
- Mermaid図表対応（フローチャート・ER図）
- シンタックスハイライト（rehype-highlight）
- プレビュー拡大モーダル
- Markdownリファレンスパネル

## 2026-04-21 Wikiページ改善
- サイドバー追加（カテゴリ一覧）
- 記事検索バー
- 戻るボタン（記事詳細ページ）
