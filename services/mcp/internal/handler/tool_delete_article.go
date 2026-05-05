package handler

import (
	"context"
	"fmt"

	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
)

func (m *MCPServer) handleDeleteArticle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := request.RequireString("id")
	if err != nil {
		return nil, err
	}

	_, err = m.wikiClient.Delete(ctx, &pb.DeleteArticleRequest{
		Id: id,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete article: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Article deleted successfully.\nID: %s", id)), nil
}
