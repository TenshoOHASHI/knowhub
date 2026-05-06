package handler

import (
	"context"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleUpdateArticle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	req := &pb.UpdateArticleRequest{
		Id: id,
	}

	args := request.GetArguments()
	if title, ok := args["title"].(string); ok && title != "" {
		req.Title = &title
	}
	if content, ok := args["content"].(string); ok && content != "" {
		req.Content = &content
	}
	if visibility, ok := args["visibility"].(string); ok && visibility != "" {
		req.Visibility = &visibility
	}

	resp, err := m.wikiClient.Update(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update article: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Article updated successfully.\nID: %s\nTitle: %s", resp.Article.Id, resp.Article.Title)), nil
}
