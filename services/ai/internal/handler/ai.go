package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/agent"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AIHandler struct {
	pb.UnimplementedAIServiceServer
	searchEngine search.SearchEngine // デフォルト（BM25）
	llmProvider  llm.LLMProvider     // デフォルト LLM
	ollamaURL    string              // Ollama embedding 用
	ollamaModel  string              // Ollama embedding model
	wikiClient   wikiPb.WikiServicesClient
	searxngURL   string // SearXNG URL

	// キャッシュ済みナレッジグラフ
	graphMu    sync.RWMutex
	graphCache *search.GraphEngine
	graphErr   error
}

func NewAIHandler(se search.SearchEngine, llm llm.LLMProvider, ollamaURL, ollamaModel string, wikiClient wikiPb.WikiServicesClient, searxngURL string) *AIHandler {
	return &AIHandler{
		searchEngine: se,
		llmProvider:  llm,
		ollamaURL:    ollamaURL,
		ollamaModel:  ollamaModel,
		wikiClient:   wikiClient,
		searxngURL:   searxngURL,
	}
}

// ensureGraph はキャッシュがあれば返し、なければ構築する
func (h *AIHandler) ensureGraph(ctx context.Context) (*search.GraphEngine, error) {
	h.graphMu.RLock()
	if h.graphCache != nil {
		cache := h.graphCache
		h.graphMu.RUnlock()
		return cache, nil
	}
	h.graphMu.RUnlock()

	// グラフ構築は全記事 × LLM API 呼び出しが発生するため、リクエストの ctx とは独立したタイムアウトを設定
	graphCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return h.buildGraph(graphCtx)
}

// buildGraph はグラフを強制的に再構築する（refresh=true 時に使用）
func (h *AIHandler) buildGraph(ctx context.Context) (*search.GraphEngine, error) {
	h.graphMu.Lock()
	defer h.graphMu.Unlock()

	// ダブルチェック: 他のリクエストが既に構築済みか確認
	if h.graphCache != nil {
		cache := h.graphCache
		return cache, nil
	}

	slog.Info("building knowledge graph...")
	docs, err := h.fetchDocs(ctx)
	if err != nil {
		h.graphErr = err
		return nil, err
	}

	engine := search.NewGraphEngine(h.llmProvider)
	if err := engine.Index(ctx, docs); err != nil {
		h.graphErr = err
		return nil, err
	}

	h.graphCache = engine
	h.graphErr = nil
	slog.Info("knowledge graph cached",
		"entities", len(engine.GetGraph().GetEntities()),
		"relations", len(engine.GetGraph().GetRelations()))

	return engine, nil
}

// invalidateGraph はキャッシュを無効化する（次回リクエストで再構築される）
func (h *AIHandler) invalidateGraph() {
	h.graphMu.Lock()
	defer h.graphMu.Unlock()
	h.graphCache = nil
	h.graphErr = nil
	slog.Info("knowledge graph cache invalidated")
}

// fetchDocs は Wiki Service から全記事を取得して search.Document に変換する
func (h *AIHandler) fetchDocs(ctx context.Context) ([]search.Document, error) {
	articles, err := h.wikiClient.List(ctx, &wikiPb.ListArticleRequest{})
	if err != nil {
		return nil, err
	}
	docs := make([]search.Document, 0, len(articles.Article))
	for _, a := range articles.Article {
		docs = append(docs, search.Document{
			ID:         a.Id,
			Title:      a.Title,
			Content:    a.Content,
			Visibility: a.Visibility,
		})
	}
	return docs, nil
}

// searchWithEngine は指定エンジンでインデックス構築 → 検索を実行する
func (h *AIHandler) searchWithEngine(ctx context.Context, se search.SearchEngine, query string, limit int) ([]search.SearchResult, error) {
	slog.Info("searchWithEngine called",
		"engine_type", fmt.Sprintf("%T", se),
		"query", query,
		"limit", limit,
	)

	docs, err := h.fetchDocs(ctx)
	if err != nil {
		slog.Error("failed to fetch articles from wiki service", "error", err)
		return nil, status.Error(codes.Internal, "failed to fetch articles from wiki service")
	}

	slog.Info("fetched articles for search", "num_docs", len(docs))

	if err := se.Index(ctx, docs); err != nil {
		slog.Error("failed to build search index", "error", err)
		return nil, status.Error(codes.Internal, "failed to build search index")
	}

	results, err := se.Search(ctx, query, limit)
	if err != nil {
		slog.Error("search failed", "error", err)
		return nil, status.Error(codes.Internal, "search failed")
	}

	slog.Info("search completed", "results_count", len(results))

	return results, nil
}

// SearchArticles は検索エンジンで関連記事を検索する
func (h *AIHandler) SearchArticles(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	results, err := h.searchWithEngine(ctx, h.searchEngine, req.Query, limit)
	if err != nil {
		return nil, err
	}

	pbResults := make([]*pb.SearchResult, 0, len(results))
	for _, r := range results {
		pbResults = append(pbResults, &pb.SearchResult{
			ArticleId:      r.ArticleID,
			Title:          r.Title,
			Content:        r.Context,
			RelevanceScore: r.RelevanceScore,
		})
	}

	return &pb.SearchResponse{Results: pbResults}, nil
}

// SummarizeArticle は指定記事を LLM で要約する
func (h *AIHandler) SummarizeArticle(ctx context.Context, req *pb.SummarizeRequest) (*pb.SummarizeResponse, error) {
	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}

	resp, err := h.wikiClient.Get(ctx, &wikiPb.GetArticleRequest{Id: req.ArticleId})
	if err != nil {
		slog.Error("failed to get article from wiki service", "error", err, "article_id", req.ArticleId)
		return nil, status.Error(codes.NotFound, "article not found")
	}

	article := resp.Article
	prompt := fmt.Sprintf(
		"以下の技術記事を日本語で簡潔に要約してください。\n\nタイトル: %s\n\n内容:\n%s",
		article.Title,
		article.Content,
	)

	summary, err := h.llmProvider.Generate(ctx, prompt)
	if err != nil {
		slog.Error("LLM summarization failed", "error", err)
		return nil, status.Error(codes.Internal, "summarization failed")
	}

	return &pb.SummarizeResponse{Summary: summary}, nil
}

// AskQuestion は RAG で記事内容に基づいて回答する
func (h *AIHandler) AskQuestion(ctx context.Context, req *pb.QuestionRequest) (*pb.QuestionResponse, error) {
	if req.Question == "" {
		return nil, status.Error(codes.InvalidArgument, "question is required")
	}

	slog.Info("RAG AskQuestion called",
		"question", req.Question,
		"model", req.Model,
		"search_engine", req.SearchEngine,
		"api_key_provided", req.ApiKey != "",
	)

	// LLM プロバイダー選択（リクエストで動的切替）
	provider := h.llmProvider
	if req.Model != "" {
		provider = llm.NewProvider(h.ollamaURL, req.Model, req.ApiKey)
	}

	// Embedding プロバイダー生成（model + apiKey から自動判定）
	embedder := embedding.NewProvider(h.ollamaURL, h.ollamaModel, req.ApiKey, req.Model)

	// 検索エンジン選択（search.SelectEngine ファクトリで動的切替）
	engine := search.SelectEngine(req.SearchEngine, provider, embedder)
	slog.Info("RAG search engine selected",
		"requested_engine", req.SearchEngine,
		"engine_type", fmt.Sprintf("%T", engine),
	)

	// Step 1: 関連記事を検索
	var results []search.SearchResult
	if req.SearchEngine == "graph" {
		// Graph RAG はキャッシュ済みグラフを使う（毎回 Index すると全記事 × LLM 呼び出しでタイムアウトする）
		graphEngine, err := h.ensureGraph(ctx)
		if err != nil {
			slog.Error("failed to get cached graph", "error", err)
			return nil, status.Error(codes.Internal, "failed to build knowledge graph")
		}
		results, err = graphEngine.Search(ctx, req.Question, 5)
		if err != nil {
			slog.Error("graph search failed", "error", err)
			return nil, status.Error(codes.Internal, "graph search failed")
		}
	} else {
		var err error
		results, err = h.searchWithEngine(ctx, engine, req.Question, 5)
		if err != nil {
			return nil, err
		}
	}

	if len(results) == 0 {
		// 検索で何も見つからない場合もLLMには回答させる。
		// ただし参照元は空にし、「参考:」として一般知識で補足する流れにする。
		results = []search.SearchResult{}
	}

	// デバッグログ: 検索結果とフィルタリング
	slog.Info("RAG search results",
		"query", req.Question,
		"search_engine", req.SearchEngine,
		"total_results", len(results),
	)
	for i, r := range results {
		slog.Info("RAG search result",
			"rank", i+1,
			"article_id", r.ArticleID,
			"title", r.Title,
			"score", r.RelevanceScore,
		)
	}

	relevantResults := filterRAGResults(req.SearchEngine, results)

	slog.Info("RAG filtered results",
		"query", req.Question,
		"search_engine", req.SearchEngine,
		"filtered_count", len(relevantResults),
		"threshold", ragSourceThreshold(req.SearchEngine),
	)

	// Step 2: コンテキストを構築
	var contextBuilder strings.Builder
	sources := make([]*pb.Source, 0, len(relevantResults))

	for _, r := range relevantResults {
		articleResp, err := h.wikiClient.Get(ctx, &wikiPb.GetArticleRequest{Id: r.ArticleID})
		if err != nil {
			slog.Error("failed to get article for RAG", "error", err, "article_id", r.ArticleID)
			continue
		}
		contextBuilder.WriteString(fmt.Sprintf("## %s\n%s\n\n", articleResp.Article.Title, articleResp.Article.Content))
		sources = append(sources, &pb.Source{
			ArticleId:      r.ArticleID,
			Title:          articleResp.Article.Title,
			RelevanceScore: r.RelevanceScore,
		})
		slog.Info("RAG article added to context",
			"article_id", r.ArticleID,
			"title", articleResp.Article.Title,
			"content_length", len(articleResp.Article.Content),
		)
	}

	contextText := contextBuilder.String()
	slog.Info("RAG context built",
		"context_length", len(contextText),
		"sources_count", len(sources),
		"context_is_empty", strings.TrimSpace(contextText) == "",
		"context_preview", string([]rune(contextText)[:min(500, len([]rune(contextText)))]),
	)
	if strings.TrimSpace(contextText) == "" {
		contextText = "関連する記事は見つかりませんでした。"
		slog.Warn("RAG context is empty, using fallback message")
	}

	// Step 3: RAG プロンプトで LLM に回答生成
	systemPrompt := "あなたは技術ナレッジベースのアシスタントです。" +
		"以下のコンテキストを参考にして回答してください。" +
		"コンテキストに記載されている記事の内容を、質問に対する回答として説明してください。" +
		"提供されたコンテキストを使って質問に答えてください。" +
		"【重要ルール】" +
		"- コンテキストに「## 」で始まる記事タイトルが含まれている場合は、必ずその記事の内容に基づいて回答してください。" +
		"- 記事が提供されているのに「関連する情報がありません」と答えるのはやめてください。" +
		"- コンテキストに「関連する記事は見つかりませんでした」としか書かれていない場合のみ、同じ内容を答えてください。" +
		"【重要】外部リンクやURLを含む回答はしないでください。Wiki内の記事のみを参照してください。"

	// Graph RAGの場合: 質問に直接関連する記事のみを使用
	if req.SearchEngine == "graph" {
		systemPrompt += " " +
			"【重要】今回はGraph RAG（ナレッジグラフ）を使った検索です。" +
			"コンテキストに含まれる記事のうち、質問に直接関連するもののみを使用して回答を構築してください。" +
			"質問と明らかに関係のないトピック（例：質問が「gRPC」なのに「MCPサーバ」や「認証」など）については、言及しないでください。" +
			"回答は简潔に、質問に対する直接的な回答に集中してください。"
	}

	messages := []llm.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"コンテキスト:\n%s\n\n質問: %s",
				contextText,
				req.Question,
			),
		},
	}

	// LLMに回答する
	answer, err := provider.Chat(ctx, messages)
	if err != nil {
		slog.Error("LLM chat failed", "error", err)
		// HTTPErrorの場合は詳細なエラー情報を返す
		if httpErr, ok := err.(*llm.HTTPError); ok {
			// レートリミットエラー
			if httpErr.IsRateLimit() {
				return nil, status.Error(codes.ResourceExhausted, httpErr.Body)
			}
			// 認証エラー
			if httpErr.IsUnauthorized() {
				return nil, status.Error(codes.Unauthenticated, httpErr.Body)
			}
			// その他のHTTPエラー
			return nil, status.Error(codes.Internal, httpErr.Body)
		}
		return nil, status.Error(codes.Internal, "failed to generate answer")
	}

	slog.Info("RAG LLM answer generated",
		"query", req.Question,
		"answer_length", len(answer),
		"answer_preview", string([]rune(answer)[:min(100, len([]rune(answer)))]),
		"sources_count_before_filter", len(sources),
	)

	// 保険として、関連性がない記事は、参照リンクを空にする
	// TODO: 一時的に無効化してデバッグ
	/*
		if answerIndicatesNoRelevantContext(answer) {
			slog.Warn("RAG LLM indicated no relevant context, clearing sources",
				"answer", answer,
			)
			sources = []*pb.Source{}
		}
	*/

	slog.Info("RAG final response",
		"sources_count", len(sources),
	)

	// 後処理: 回答フォーマットを統一（LLMモデルに依存しないように）
	hasContext := len(sources) > 0
	formattedAnswer := formatRAGAnswer(answer, hasContext)

	return &pb.QuestionResponse{
		Answer:  formattedAnswer,
		Sources: sources,
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// formatRAGAnswer はRAG回答のフォーマットを統一する（LLMモデルに依存しない）
func formatRAGAnswer(answer string, hasContext bool) string {
	if hasContext {
		// コンテキストがある場合：先頭に「Wikiの情報を参考に回答します。」を追加
		// ただし、既に含まれている場合は重複を避ける
		prefix := "### Wikiの情報を参考に回答します。"
		if !strings.HasPrefix(answer, prefix) && !strings.HasPrefix(answer, "Wikiの情報を参考に回答します。") {
			return prefix + "\n\n" + answer
		}
		return answer
	}
	// コンテキストがない場合：回答をそのまま返す
	return answer
}

// 閾値を用意し、関連性が低いスコアは参照先から取り除く
func filterRAGResults(engineName string, results []search.SearchResult) []search.SearchResult {
	threshold := ragSourceThreshold(engineName)
	filtered := make([]search.SearchResult, 0, len(results))
	for _, r := range results {
		if r.RelevanceScore >= threshold {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// ragSourceThreshold は「RAGの根拠として表示してよい最低スコア」を返す。
// 検索エンジンごとにスコアの尺度が違うため、同じ閾値にはしない。
//
// スコアの範囲（目安）:
//   - BM25:    0〜3程度    （キーワード一致度）
//   - Vector:  0〜1        （コサイン類似度）
//   - Hybrid:  0〜1        （正規化済みスコア）
//   - Graph:   0〜N        （グラフに基づく関連スコア）
//   - TF-IDF:  0〜0.5程度  （TF-IDF重み）
//
// フロントエンドでの表示:
//
//	スコアの尺度がエンジンごとに異なるため、フロントエンド側で
//	パーセンテージに正規化して表示している（ChatInterface.tsx参照）。
func ragSourceThreshold(engineName string) float64 {
	switch engineName {
	case "vector":
		return 0.30 // Vector検索は閾値を下げる（一般的な単語でもマッチさせる）
	case "hybrid":
		return 0.50 // ハイブリッドスコアが50%以上の記事のみを参照
	case "graph":
		return 15.0 // Graph RAGは閾値を上げて、強く関連する記事のみを参照
	case "tfidf":
		return 0.08
	case "bm25", "":
		return 0.01 // BM25はキーワード一致なので、閾値をほぼ0にする
	default:
		return 0.0
	}
}

// answerIndicatesNoRelevantContext は、LLM自身が「コンテキストに関連情報がない」と
// 判断した回答かを見て、無関係な参照リンクを出さないための最終ガード。
func answerIndicatesNoRelevantContext(answer string) bool {
	phrases := []string{
		"コンテキストには関連情報がありません",
		"提供されたコンテキストには関連情報がありません",
		"関連情報がありません",
		"関連する情報はありません",
		"関連する記事は見つかりません",
		"ナレッジベースには関連する情報がありません。別の質問をお願いします。",
	}
	for _, phrase := range phrases {
		if strings.Contains(answer, phrase) {
			return true
		}
	}
	return false
}

// GetKnowledgeGraph はナレッジグラフの全データ（エンティティ・リレーション）を返す
func (h *AIHandler) GetKnowledgeGraph(ctx context.Context, req *pb.GetKnowledgeGraphRequest) (*pb.GetKnowledgeGraphResponse, error) {
	graphEngine, err := h.ensureGraph(ctx)
	if err != nil {
		slog.Error("failed to build knowledge graph", "error", err)
		return nil, status.Error(codes.Internal, "failed to build knowledge graph")
	}

	graph := graphEngine.GetGraph()

	pbEntities := make([]*pb.EntityNode, 0, len(graph.GetEntities()))
	for _, e := range graph.GetEntities() {
		pbEntities = append(pbEntities, &pb.EntityNode{
			Id:         e.ID,
			Name:       e.Name,
			Type:       e.Type,
			ArticleIds: e.ArticleIDs,
		})
	}

	pbRelations := make([]*pb.RelationEdge, 0, len(graph.GetRelations()))
	for _, r := range graph.GetRelations() {
		pbRelations = append(pbRelations, &pb.RelationEdge{
			Source: r.Source,
			Target: r.Target,
			Label:  r.Label,
		})
	}

	return &pb.GetKnowledgeGraphResponse{
		Entities:  pbEntities,
		Relations: pbRelations,
	}, nil
}

// GetRelatedArticles は指定記事の関連記事を BFS で収集して返す
func (h *AIHandler) GetRelatedArticles(ctx context.Context, req *pb.GetRelatedArticlesRequest) (*pb.GetRelatedArticlesResponse, error) {
	if req.ArticleId == "" {
		return nil, status.Error(codes.InvalidArgument, "article_id is required")
	}

	graphEngine, err := h.ensureGraph(ctx)
	if err != nil {
		slog.Error("failed to build knowledge graph", "error", err)
		return nil, status.Error(codes.Internal, "failed to build knowledge graph")
	}

	graph := graphEngine.GetGraph()

	maxHops := int(req.MaxHops)
	if maxHops <= 0 {
		maxHops = 2
	}

	relatedScores := graph.GetRelatedArticleIDs(req.ArticleId, maxHops)

	// docID → Document の map
	docMap := make(map[string]search.Document)
	for _, doc := range graphEngine.GetDocs() {
		docMap[doc.ID] = doc
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 5
	}

	results := make([]*pb.SearchResult, 0, len(relatedScores))
	for aid, score := range relatedScores {
		doc, ok := docMap[aid]
		if !ok {
			continue
		}
		snippet := doc.Content
		if len(snippet) > 200 {
			runes := []rune(snippet)
			if len(runes) > 200 {
				snippet = string(runes[:200]) + "..."
			}
		}
		results = append(results, &pb.SearchResult{
			ArticleId:      aid,
			Title:          doc.Title,
			Content:        snippet,
			RelevanceScore: float64(score),
		})
	}

	// スコア降順ソート
	sort.Slice(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	if limit < len(results) {
		results = results[:limit]
	}

	return &pb.GetRelatedArticlesResponse{Results: results}, nil
}

// InvalidateGraphCache はキャッシュされたグラフを無効化する（次回アクセス時に再構築される）
func (h *AIHandler) InvalidateGraphCache(ctx context.Context, req *pb.InvalidateGraphCacheRequest) (*pb.InvalidateGraphCacheResponse, error) {
	h.invalidateGraph()
	return &pb.InvalidateGraphCacheResponse{}, nil
}

// AskWithAgent は ReAct Agent で自律的にツールを選択・実行して回答する
func (h *AIHandler) AskWithAgent(ctx context.Context, req *pb.AgentQuestionRequest) (*pb.AgentQuestionResponse, error) {
	if req.Question == "" {
		return nil, status.Error(codes.InvalidArgument, "question is required")
	}

	// LLM プロバイダー選択
	provider := h.llmProvider
	if req.Model != "" {
		provider = llm.NewProvider(h.ollamaURL, req.Model, req.ApiKey)
	}

	// Embedding プロバイダー生成（model + apiKey から自動判定）
	embedder := embedding.NewProvider(h.ollamaURL, h.ollamaModel, req.ApiKey, req.Model)

	// ツール構築
	var tools []agent.Tool
	tools = append(tools,
		agent.NewSearchWikiTool(h.wikiClient, provider, embedder, req.SearchEngine, h.ensureGraph),
		agent.NewReadArticleTool(h.wikiClient),
		agent.NewListArticlesTool(h.wikiClient),
	)

	if req.EnableWebSearch {
		tools = append(tools,
			agent.NewWebSearchTool(h.searxngURL),
			agent.NewReadURLTool(),
		)
	}

	callbacks := agent.NewLoggingCallbacks()
	ag := agent.NewAgent(provider, tools, 10, callbacks)

	// 会話履歴をパース
	var history []llm.Message
	if req.History != "" {
		var entries []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(req.History), &entries); err == nil {
			for _, e := range entries {
				history = append(history, llm.Message{Role: e.Role, Content: e.Content})
			}
		}
	}

	// 外部モデル（Gemini/DeepSeek/OpenAI等）→ 自律ReAct、Ollama → 固定パイプライン
	var result *agent.AgentResult
	var err error
	if isExternalModel(req.Model) {
		slog.Info("using autonomous ReAct mode", "model", req.Model)
		result, err = ag.Run(ctx, req.Question)
	} else {
		slog.Info("using pipeline mode", "model", req.Model)
		result, err = ag.RunPipeline(ctx, req.Question, history)
	}
	if err != nil {
		slog.Error("agent run failed", "error", err, "model", req.Model)
		// HTTPErrorの場合は詳細なエラー情報を返す
		if httpErr, ok := err.(*llm.HTTPError); ok {
			if httpErr.IsRateLimit() {
				return nil, status.Error(codes.ResourceExhausted, httpErr.Body)
			}
			if httpErr.IsUnauthorized() {
				return nil, status.Error(codes.Unauthenticated, httpErr.Body)
			}
			return nil, status.Error(codes.Internal, httpErr.Body)
		}
		return nil, status.Error(codes.Internal, "agent execution failed")
	}

	// proto レスポンスに変換
	pbSteps := make([]*pb.AgentStep, 0, len(result.Steps))
	for _, s := range result.Steps {
		pbSteps = append(pbSteps, &pb.AgentStep{
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionInput,
			Observation: s.Observation,
		})
	}

	pbSources := make([]*pb.AgentSource, 0, len(result.Sources))
	for _, s := range result.Sources {
		pbSources = append(pbSources, &pb.AgentSource{
			ArticleId:      s.ArticleID,
			Title:          s.Title,
			Url:            s.URL,
			RelevanceScore: s.RelevanceScore,
		})
	}

	// 関連性がない記事は、参照リンクを空にする
	if answerIndicatesNoRelevantContext(result.Answer) {
		pbSources = []*pb.AgentSource{}
	}

	return &pb.AgentQuestionResponse{
		Answer:  result.Answer,
		Steps:   pbSteps,
		Sources: pbSources,
	}, nil
}

// isExternalModel は外部LLMプロバイダー（Gemini/DeepSeek/OpenAI/GLM-5）かどうかを判定する
// llm.NewProvider のプレフィックス判定と同じロジック
func isExternalModel(model string) bool {
	return strings.HasPrefix(model, "deepseek") ||
		strings.HasPrefix(model, "gemini") ||
		strings.HasPrefix(model, "glm") ||
		strings.HasPrefix(model, "gpt")
}

// AskWithAgentStream は ReAct Agent の実行過程を gRPC server-streaming でリアルタイム送信する
func (h *AIHandler) AskWithAgentStream(req *pb.AgentQuestionRequest, stream grpc.ServerStreamingServer[pb.AgentStreamEvent]) error {
	if req.Question == "" {
		return status.Error(codes.InvalidArgument, "question is required")
	}

	// LLM プロバイダー選択
	provider := h.llmProvider
	if req.Model != "" {
		provider = llm.NewProvider(h.ollamaURL, req.Model, req.ApiKey)
	}

	// Embedding プロバイダー生成
	embedder := embedding.NewProvider(h.ollamaURL, h.ollamaModel, req.ApiKey, req.Model)

	// ツール構築
	var tools []agent.Tool
	tools = append(tools,
		agent.NewSearchWikiTool(h.wikiClient, provider, embedder, req.SearchEngine, h.ensureGraph),
		agent.NewReadArticleTool(h.wikiClient),
		agent.NewListArticlesTool(h.wikiClient),
	)

	if req.EnableWebSearch {
		tools = append(tools,
			agent.NewWebSearchTool(h.searxngURL),
			agent.NewReadURLTool(),
		)
	}

	// ストリーミングコールバック: gRPC stream.Send() でイベントを送信
	callbacks := agent.NewStreamingCallbacks(func(event *pb.AgentStreamEvent) error {
		return stream.Send(event)
	})
	ag := agent.NewAgent(provider, tools, 10, callbacks)

	// 会話履歴をパース
	var history []llm.Message
	if req.History != "" {
		var entries []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(req.History), &entries); err == nil {
			for _, e := range entries {
				history = append(history, llm.Message{Role: e.Role, Content: e.Content})
			}
		}
	}

	// Agent 実行
	var result *agent.AgentResult
	var err error
	if isExternalModel(req.Model) {
		slog.Info("using autonomous ReAct mode (streaming)", "model", req.Model)
		result, err = ag.Run(stream.Context(), req.Question)
	} else {
		slog.Info("using pipeline mode (streaming)", "model", req.Model)
		result, err = ag.RunPipeline(stream.Context(), req.Question, history)
	}
	if err != nil {
		slog.Error("agent run failed (streaming)", "error", err, "model", req.Model)
		// HTTPErrorの場合は詳細なエラー情報を返す
		var errMsg string
		if httpErr, ok := err.(*llm.HTTPError); ok {
			errMsg = httpErr.Body
		} else {
			errMsg = err.Error()
		}
		_ = stream.Send(&pb.AgentStreamEvent{
			EventType: "error",
			Error:     &pb.AgentErrorEvent{Message: errMsg},
		})
		// gRPCステータスコードを適切に返す
		if httpErr, ok := err.(*llm.HTTPError); ok {
			if httpErr.IsRateLimit() {
				return status.Error(codes.ResourceExhausted, httpErr.Body)
			}
			if httpErr.IsUnauthorized() {
				return status.Error(codes.Unauthenticated, httpErr.Body)
			}
			return status.Error(codes.Internal, httpErr.Body)
		}
		return status.Error(codes.Internal, "agent execution failed")
	}

	// proto ステップに変換
	pbSteps := make([]*pb.AgentStep, 0, len(result.Steps))
	for _, s := range result.Steps {
		pbSteps = append(pbSteps, &pb.AgentStep{
			Thought:     s.Thought,
			Action:      s.Action,
			ActionInput: s.ActionInput,
			Observation: s.Observation,
		})
	}

	pbSources := make([]*pb.AgentSource, 0, len(result.Sources))
	for _, s := range result.Sources {
		pbSources = append(pbSources, &pb.AgentSource{
			ArticleId:      s.ArticleID,
			Title:          s.Title,
			Url:            s.URL,
			RelevanceScore: s.RelevanceScore,
		})
	}

	// 関連性がない記事は、参照リンクを空にする
	if answerIndicatesNoRelevantContext(result.Answer) {
		pbSources = []*pb.AgentSource{}
	}

	// final_answer イベントを送信
	if err := stream.Send(&pb.AgentStreamEvent{
		EventType: "final_answer",
		Final: &pb.AgentFinalEvent{
			Answer:  result.Answer,
			Steps:   pbSteps,
			Sources: pbSources,
		},
	}); err != nil {
		return err
	}

	return nil
}
