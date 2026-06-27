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
			mcp.WithDescription("列出数据目录下所有已下载的作者及其文章数（按 canonical id 去重）"),
			mcp.WithOutputSchema[authorsOut](),
		),
		handleListAuthors(dataDir),
	)

	// 注册 list_posts Tool
	s.AddTool(
		mcp.NewTool("list_posts",
			mcp.WithDescription("列出指定作者下的文章（按 canonical id 去重），支持作品集/标题/日期过滤与分页，返回精简元信息"),
			mcp.WithString("author",
				mcp.Required(),
				mcp.Description("作者的 URL slug（即目录名）"),
			),
			mcp.WithString("collection",
				mcp.Description("仅返回属于该作品集的文章（可选）"),
			),
			mcp.WithString("title_contains",
				mcp.Description("标题子串过滤，大小写不敏感（可选）"),
			),
			mcp.WithString("date_from",
				mcp.Description("起始日期 YYYY-MM-DD，含（可选）"),
			),
			mcp.WithString("date_to",
				mcp.Description("结束日期 YYYY-MM-DD，含（可选）"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回上限，默认 50"),
			),
			mcp.WithNumber("offset",
				mcp.Description("偏移，默认 0"),
			),
			mcp.WithOutputSchema[postsOut](),
		),
		handleListPosts(dataDir),
	)

	// 注册 read_post Tool（输出可能是文章正文或消歧列表，故不声明 OutputSchema）
	s.AddTool(
		mcp.NewTool("read_post",
			mcp.WithDescription("读取一篇文章的完整内容。优先级：id > url > path > author+title"),
			mcp.WithString("id",
				mcp.Description("canonical id（afdian 文章 hash），最精确"),
			),
			mcp.WithString("url",
				mcp.Description("afdian 文章 URL 或裸 hash"),
			),
			mcp.WithString("path",
				mcp.Description("文章的相对路径（相对于数据目录）"),
			),
			mcp.WithString("author",
				mcp.Description("作者名（与 title 配合使用）"),
			),
			mcp.WithString("title",
				mcp.Description("文章标题关键词（模糊匹配；去重后多义时返回带 id 的候选列表）"),
			),
		),
		handleReadPost(dataDir),
	)

	// 注册 search Tool
	s.AddTool(
		mcp.NewTool("search",
			mcp.WithDescription("在已下载文档中全文搜索。多概念请用空格分隔（如 `信念 执念 信仰`），默认 AND；\"短语\" 精确匹配。结果按 canonical id 去重并按相关性排序，文档级聚合。"),
			mcp.WithString("query",
				mcp.Required(),
				mcp.Description("搜索关键词；空格/标点切分为多个词，默认全部命中(AND)；用英文双引号包裹的为短语精确匹配"),
			),
			mcp.WithString("match",
				mcp.Description("all=全部命中(AND，默认)，any=任一命中(OR)"),
			),
			mcp.WithString("author",
				mcp.Description("限定作者（可选）"),
			),
			mcp.WithString("collection",
				mcp.Description("限定作品集（可选）"),
			),
			mcp.WithNumber("limit",
				mcp.Description("文档级返回上限，默认 10"),
			),
			mcp.WithNumber("offset",
				mcp.Description("文档级偏移，默认 0"),
			),
			mcp.WithNumber("per_doc_hits",
				mcp.Description("每篇文档返回的片段上限，默认 3"),
			),
			mcp.WithOutputSchema[searchOut](),
		),
		handleSearch(dataDir),
	)

	// 注册 list_comments Tool
	s.AddTool(
		mcp.NewTool("list_comments",
			mcp.WithDescription("列出一篇文章的评论（热评 + 普通评论），支持按评论者过滤与分页。定位参数 id/url/path 三选一"),
			mcp.WithString("id",
				mcp.Description("canonical id（afdian 文章 hash）"),
			),
			mcp.WithString("url",
				mcp.Description("afdian 文章 URL 或裸 hash"),
			),
			mcp.WithString("path",
				mcp.Description("文章的相对路径（相对于数据目录）"),
			),
			mcp.WithString("commenter",
				mcp.Description("评论者名子串过滤，大小写不敏感（可选）"),
			),
			mcp.WithBoolean("hot_only",
				mcp.Description("仅返回热评，默认 false"),
			),
			mcp.WithNumber("limit",
				mcp.Description("返回上限，默认 50"),
			),
			mcp.WithNumber("offset",
				mcp.Description("偏移，默认 0"),
			),
			mcp.WithOutputSchema[commentsOut](),
		),
		handleListComments(dataDir),
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
