package handler

import (
	"context"
	"fmt"
	"strings"

	aiPb "github.com/TenshoOHASHI/knowhub/proto/ai"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleSearchArticles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return nil, err
	}

	limit := int32(10)
	args := request.GetArguments()
	if l, ok := args["limit"].(float64); ok {
		limit = int32(l)
	}

	resp, err := m.aiClient.SearchArticles(ctx, &aiPb.SearchRequest{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to search articles: %v", err)), nil
	}

	if len(resp.Results) == 0 {
		return mcp.NewToolResultText("No articles found matching your query."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d articles:\n\n", len(resp.Results)))
	for i, r := range resp.Results {
		sb.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n   Score: %.4f\n   %s\n\n", i+1, r.Title, r.ArticleId, r.RelevanceScore, truncate(r.Content, 200)))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
