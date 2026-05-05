package handler

import (
	"context"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleCreateArticle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, err := request.RequireString("title")
	if err != nil {
		return nil, err
	}
	content, err := request.RequireString("content")
	if err != nil {
		return nil, err
	}

	categoryID := request.GetString("category_id", "")
	visibility := request.GetString("visibility", "public")

	resp, err := m.wikiClient.Create(ctx, &pb.CreateArticleRequest{
		Title:      title,
		Content:    content,
		CategoryId: categoryID,
		Visibility: visibility,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create article: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Article created successfully.\nID: %s\nTitle: %s", resp.Article.Id, resp.Article.Title)), nil
}
