package handler

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleListArticles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resp, err := m.wikiClient.List(ctx, &pb.ListArticleRequest{})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list articles: %v", err)), nil
	}

	if len(resp.Article) == 0 {
		return mcp.NewToolResultText("No articles found."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Total articles: %d\n\n", len(resp.Article)))
	for i, a := range resp.Article {
		sb.WriteString(fmt.Sprintf("%d. **%s** (ID: %s)\n   Visibility: %s\n\n", i+1, a.Title, a.Id, a.Visibility))
	}

	return mcp.NewToolResultText(sb.String()), nil
}
