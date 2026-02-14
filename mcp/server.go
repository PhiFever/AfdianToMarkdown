package mcp

import (
	"context"
	"golang.org/x/exp/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer 创建并配置 MCP Server，注册所有 Tool
func NewServer(dataDir string, version string) *server.MCPServer {
	s := server.NewMCPServer(
		"AfdianToMarkdown",
		version,
	)

	// 注册 list_authors Tool
	s.AddTool(
		mcp.NewTool("list_authors",
			mcp.WithDescription("列出数据目录下所有已下载的作者"),
		),
		handleListAuthors(dataDir),
	)

	// 注册 list_posts Tool
	s.AddTool(
		mcp.NewTool("list_posts",
			mcp.WithDescription("列出指定作者下的所有文章，按动态和作品集分组"),
			mcp.WithString("author",
				mcp.Required(),
				mcp.Description("作者的 URL slug（即目录名）"),
			),
		),
		handleListPosts(dataDir),
	)

	// 注册 read_post Tool
	s.AddTool(
		mcp.NewTool("read_post",
			mcp.WithDescription("读取指定文章的完整 Markdown 内容"),
			mcp.WithString("path",
				mcp.Description("文章的相对路径（相对于数据目录）"),
			),
			mcp.WithString("author",
				mcp.Description("作者名（与 title 配合使用）"),
			),
			mcp.WithString("title",
				mcp.Description("文章标题关键词（模糊匹配）"),
			),
		),
		handleReadPost(dataDir),
	)

	// 注册 search Tool
	s.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("在已下载文档中全文搜索关键词"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("搜索关键词"),
			),
			mcp.WithString("author",
				mcp.Description("限定搜索范围的作者名（可选）"),
			),
		),
		handleSearch(dataDir),
	)

	return s
}

// Serve 以 stdio 传输模式启动 MCP Server
func Serve(s *server.MCPServer) error {
	return server.ServeStdio(s)
}

// ServeHTTP 以 HTTP Streamable 传输模式启动 MCP Server
func ServeHTTP(s *server.MCPServer, addr string) error {
	httpServer := server.NewStreamableHTTPServer(s)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("MCP HTTP Server 正在监听", "addr", addr, "endpoint", "/mcp")
		errCh <- httpServer.Start(addr)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("正在关闭 MCP HTTP Server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	}
}
