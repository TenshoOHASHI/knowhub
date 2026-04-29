package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	pb "github.com/TenshoOHASHI/knowhub/proto/ai"
	wikiPb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/llm"
	"github.com/TenshoOHASHI/knowhub/services/ai/internal/search"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AIHandler struct {
	pb.UnimplementedAIServiceServer
	searchEngine search.SearchEngine
	llmProvider  llm.LLMProvider
	wikiClient   wikiPb.WikiServicesClient // Wiki Service の gRPCサーバーと通信するクライアント

}

func NewAIHandler(se search.SearchEngine, llm llm.LLMProvider, wikiClient wikiPb.WikiServicesClient) *AIHandler {
	return &AIHandler{
		searchEngine: se,
		llmProvider:  llm,
		wikiClient:   wikiClient,
	}
}

// SearchArticles は TF-IDF / BM25 で関連記事を検索する
func (h *AIHandler) SearchArticles(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	// Wiki Service から全記事を取得してインデックスを構築
	articles, err := h.wikiClient.List(ctx, &wikiPb.ListArticleRequest{})
	if err != nil {
		slog.Error("failed to fetch articles from wiki service", "error", err)
		return nil, status.Error(codes.Internal, "failed to fetch articles from wiki service")
	}

	// 記事を SearchEngine のドキュメントに変換してインデックス構築
	docs := make([]search.Document, 0, len(articles.Article))
	for _, a := range articles.Article {
		docs = append(docs, search.Document{
			ID:      a.Id,
			Title:   a.Title,
			Content: a.Content,
		})
	}

	if err := h.searchEngine.Index(ctx, docs); err != nil {
		slog.Error("failed to build search index", "error", err)
		return nil, status.Error(codes.Internal, "failed to build search index")
	}

	// 検索実行
	results, err := h.searchEngine.Search(ctx, req.Query, limit)
	if err != nil {
		slog.Error("search failed", "error", err)
		return nil, status.Error(codes.Internal, "search failed")
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

	// Wiki Service から記事を取得
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

	// Step 1: 関連記事を検索
	searchResp, err := h.SearchArticles(ctx, &pb.SearchRequest{
		Query: req.Question,
		Limit: 5,
	})
	if err != nil {
		return nil, err
	}

	if len(searchResp.Results) == 0 {
		return &pb.QuestionResponse{
			Answer:  "関連する記事が見つかりませんでした。",
			Sources: []string{},
		}, nil
	}

	// Step 2: コンテキストを構築
	var contextBuilder strings.Builder
	sources := make([]string, 0, len(searchResp.Results))

	for _, r := range searchResp.Results {
		// スニペット
		// contextBuilder.WriteString(fmt.Sprintf("## %s\n%s\n\n", r.Title, r.Content))
		// 全文を検索
		articleResp, err := h.wikiClient.Get(ctx, &wikiPb.GetArticleRequest{Id: r.ArticleId})
		if err != nil {
			slog.Error("failed to get article fro RAG", "error", err, "article_id", r.ArticleId)
			continue
		}
		contextBuilder.WriteString(fmt.Sprintf("## %s\n%s\n\n", articleResp.Article.Title, articleResp.Article.Content))
		sources = append(sources, r.ArticleId)
	}

	// Step 3: RAG プロンプトで LLM に回答生成
	messages := []llm.Message{
		{
			Role: "system",
			Content: "あなたは技術ナレッジベースのアシスタントです。" +
				"以下のコンテキストを参考にして回答してください。" +
				"コンテキストに情報がある場合はそれを優先し、" +
				// "ない場合はあなたの知識で補足してください。" +
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

	answer, err := h.llmProvider.Chat(ctx, messages)
	if err != nil {
		slog.Error("LLM chat failed", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate answer")
	}

	return &pb.QuestionResponse{
		Answer:  answer,
		Sources: sources,
	}, nil
}
