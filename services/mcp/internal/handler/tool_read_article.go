package handler

import (
	"context"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleReadArticle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	articleID, err := request.RequireString("article_id")
	if err != nil {
		return nil, err
	}

	resp, err := m.wikiClient.Get(ctx, &pb.GetArticleRequest{
		Id: articleID,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read article: %v", err)), nil
	}

	a := resp.Article
	result := fmt.Sprintf("# %s\n\nID: %s\nVisibility: %s\nCategory: %s\n\n---\n\n%s", a.Title, a.Id, a.Visibility, a.CategoryId, a.Content)

	return mcp.NewToolResultText(result), nil
}
