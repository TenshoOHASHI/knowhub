package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/agent"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/embedding"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
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

	return h.buildGraph(ctx)
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
			ID:      a.Id,
			Title:   a.Title,
			Content: a.Content,
		})
	}
	return docs, nil
}

// searchWithEngine は指定エンジンでインデックス構築 → 検索を実行する
func (h *AIHandler) searchWithEngine(ctx context.Context, se search.SearchEngine, query string, limit int) ([]search.SearchResult, error) {
	docs, err := h.fetchDocs(ctx)
	if err != nil {
		slog.Error("failed to fetch articles from wiki service", "error", err)
		return nil, status.Error(codes.Internal, "failed to fetch articles from wiki service")
	}
	if err := se.Index(ctx, docs); err != nil {
		slog.Error("failed to build search index", "error", err)
		return nil, status.Error(codes.Internal, "failed to build search index")
	}
	results, err := se.Search(ctx, query, limit)
	if err != nil {
		slog.Error("search failed", "error", err)
		return nil, status.Error(codes.Internal, "search failed")
	}
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

	// LLM プロバイダー選択（リクエストで動的切替）
	provider := h.llmProvider
	if req.Model != "" {
		provider = llm.NewProvider(req.Model, req.ApiKey)
	}

	// Embedding プロバイダー生成（model + apiKey から自動判定）
	embedder := embedding.NewProvider(h.ollamaURL, h.ollamaModel, req.ApiKey, req.Model)

	// 検索エンジン選択（search.SelectEngine ファクトリで動的切替）
	engine := search.SelectEngine(req.SearchEngine, provider, embedder)

	// Step 1: 関連記事を検索
	results, err := h.searchWithEngine(ctx, engine, req.Question, 5)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &pb.QuestionResponse{
			Answer:  "関連する記事が見つかりませんでした。",
			Sources: []*pb.Source{},
		}, nil
	}

	// Step 2: コンテキストを構築
	var contextBuilder strings.Builder
	sources := make([]*pb.Source, 0, len(results))

	for _, r := range results {
		articleResp, err := h.wikiClient.Get(ctx, &wikiPb.GetArticleRequest{Id: r.ArticleID})
		if err != nil {
			slog.Error("failed to get article for RAG", "error", err, "article_id", r.ArticleID)
			continue
		}
		contextBuilder.WriteString(fmt.Sprintf("## %s\n%s\n\n", articleResp.Article.Title, articleResp.Article.Content))
		sources = append(sources, &pb.Source{
			ArticleId: r.ArticleID,
			Title:     articleResp.Article.Title,
		})
	}

	// Step 3: RAG プロンプトで LLM に回答生成
	messages := []llm.Message{
		{
			Role: "system",
			Content: "あなたは技術ナレッジベースのアシスタントです。" +
				"以下のコンテキストを参考にして回答してください。" +
				"コンテキストに情報がある場合はそれを優先し、" +
				"もしその情報に追加した方がいい情報があれば、あなたの知識で補足してください。" +
				"ただし、あなたの知識で補足する場合は「参考:」と明記してください。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"コンテキスト:\n%s\n\n質問: %s",
				contextBuilder.String(),
				req.Question,
			),
		},
	}

	answer, err := provider.Chat(ctx, messages)
	if err != nil {
		slog.Error("LLM chat failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate answer")
	}

	return &pb.QuestionResponse{
		Answer:  answer,
		Sources: sources,
	}, nil
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
		provider = llm.NewProvider(req.Model, req.ApiKey)
	}

	// Embedding プロバイダー生成（model + apiKey から自動判定）
	embedder := embedding.NewProvider(h.ollamaURL, h.ollamaModel, req.ApiKey, req.Model)

	// ツール構築
	var tools []agent.Tool
	tools = append(tools,
		agent.NewSearchWikiTool(h.wikiClient, provider, embedder, req.SearchEngine),
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
			ArticleId: s.ArticleID,
			Title:     s.Title,
			Url:       s.URL,
		})
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
