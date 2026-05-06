package handler

import (
	"context"
	"fmt"
	"strings"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleArticleIndexResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	resp, err := m.wikiClient.List(ctx, &pb.ListArticleRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list articles: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("# Article Index\n\n")

	if len(resp.Article) == 0 {
		sb.WriteString("No articles found.\n")
	} else {
		for i, a := range resp.Article {
			sb.WriteString(fmt.Sprintf("%d. [%s](wiki://articles/%s)\n", i+1, a.Title, a.Id))
		}
	}

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      "wiki://articles/index",
			MIMEType: "text/markdown",
			Text:     sb.String(),
		},
	}, nil
}
