package handler

import (
	aiPb "github.com/TenshoOHASHI/knowhub/proto/ai"
	pb "github.com/TenshoOHASHI/knowhub/proto/wiki"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the MCP server with gRPC clients.
type MCPServer struct {
	wikiClient pb.WikiServicesClient
	aiClient   aiPb.AIServiceClient
	server     *server.MCPServer
}

// NewMCPServer creates a new MCP server with all tools and resources registered.
func NewMCPServer(wikiClient pb.WikiServicesClient, aiClient aiPb.AIServiceClient) *MCPServer {
	s := server.NewMCPServer(
		"TenHub Wiki",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// 翻訳機のインスタンスを作成（アダプターパターン）
	mcpServer := &MCPServer{
		wikiClient: wikiClient,
		aiClient:   aiClient,
		server:     s,
	}

	// 使用可能なツールの一覧を登録
	mcpServer.registerTools()
	// 資料を閲覧する窓口を登録
	mcpServer.registerResources()

	return mcpServer
}

// GetServer returns the underlying MCP server.
func (m *MCPServer) GetServer() *server.MCPServer {
	return m.server
}

func (m *MCPServer) registerTools() {
	// create_article
	createTool := mcp.NewTool("create_article",
		mcp.WithDescription("Create a new wiki article"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Article title"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Article content in Markdown"),
		),
		mcp.WithString("category_id",
			mcp.Description("Category ID to assign"),
		),
		mcp.WithString("visibility",
			mcp.Description("Visibility setting: public or locked"),
		),
	)
	m.server.AddTool(createTool, m.handleCreateArticle)

	// search_articles
	searchTool := mcp.NewTool("search_articles",
		mcp.WithDescription("Search wiki articles using AI-powered search"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results"),
		),
	)
	m.server.AddTool(searchTool, m.handleSearchArticles)

	// list_articles
	listTool := mcp.NewTool("list_articles",
		mcp.WithDescription("List all wiki articles"),
	)
	m.server.AddTool(listTool, m.handleListArticles)

	// read_article
	readTool := mcp.NewTool("read_article",
		mcp.WithDescription("Read a specific wiki article by ID"),
		mcp.WithString("article_id",
			mcp.Required(),
			mcp.Description("Article ID"),
		),
	)
	m.server.AddTool(readTool, m.handleReadArticle)

	// update_article
	updateTool := mcp.NewTool("update_article",
		mcp.WithDescription("Update an existing wiki article"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Article ID"),
		),
		mcp.WithString("title",
			mcp.Description("New title (optional)"),
		),
		mcp.WithString("content",
			mcp.Description("New content in Markdown (optional)"),
		),
		mcp.WithString("visibility",
			mcp.Description("New visibility: public or locked (optional)"),
		),
	)
	m.server.AddTool(updateTool, m.handleUpdateArticle)

	// delete_article
	deleteTool := mcp.NewTool("delete_article",
		mcp.WithDescription("Delete a wiki article by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Article ID"),
		),
	)
	m.server.AddTool(deleteTool, m.handleDeleteArticle)
}

// リストを用意（サーバ側が自動でクライアントにリスト送信、クライアントはリストのリソースをもとに、paramsを送信）
// リソースを読む時は、resources/readの規格を使用
// ツールを実行する場合は、tools/callの規格を使用
func (m *MCPServer) registerResources() {
	// Article index resource
	articleIndex := mcp.NewResource(
		"wiki://articles/index",
		"Article Index",
		mcp.WithMIMEType("text/markdown"),
	)
	//　静的リソースを登録
	m.server.AddResource(articleIndex, m.handleArticleIndexResource)

	// Individual article resource template
	articleTemplate := mcp.NewResourceTemplate(
		"wiki://articles/{id}",
		"Article Content",
		mcp.WithTemplateMIMEType("text/markdown"),
	)
	//　動的なリソースを登録
	m.server.AddResourceTemplate(articleTemplate, m.handleArticleResource)
}
