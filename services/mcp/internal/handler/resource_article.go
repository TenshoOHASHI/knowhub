package handler

import (
	"context"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleArticleResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract article ID from URI: wiki://articles/{id}
	articleID := request.Params.URI
	if len(articleID) > 20 { // "wiki://articles/" is 19 chars
		articleID = articleID[20:]
	}

	resp, err := m.wikiClient.Get(ctx, &pb.GetArticleRequest{
		Id: articleID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	a := resp.Article
	text := fmt.Sprintf("# %s\n\nID: %s\nVisibility: %s\nCategory: %s\n\n---\n\n%s", a.Title, a.Id, a.Visibility, a.CategoryId, a.Content)

	return []mcp.ResourceContents{
		&mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/markdown",
			Text:     text,
		},
	}, nil
}
