package mcp

import (
	"AfdianToMarkdown/storage"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleListAuthors 处理 list_authors Tool 调用
// 返回数据目录下所有已下载的作者列表
func handleListAuthors(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authors, err := storage.ListAuthors(dataDir)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(authors) == 0 {
			return mcp.NewToolResultText("当前没有已下载的作者。"), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("已下载的作者（共 %d 位）：\n", len(authors)))
		for _, a := range authors {
			sb.WriteString(fmt.Sprintf("- %s\n", a))
		}
		return mcp.NewToolResultText(sb.String()), nil
	}
}

// handleListPosts 处理 list_posts Tool 调用
// 返回指定作者下按动态和作品集分组的文章列表
func handleListPosts(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		author, err := request.RequireString("author")
		if err != nil {
			return mcp.NewToolResultError("请提供 author 参数"), nil
		}

		authorPosts, err := storage.ListPosts(dataDir, author)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// 检查是否有文章
		totalCount := len(authorPosts.Motions)
		for _, posts := range authorPosts.Albums {
			totalCount += len(posts)
		}
		if totalCount == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("作者 %s 下没有任何文章。", author)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("作者：%s\n", author))

		// 输出动态
		if len(authorPosts.Motions) > 0 {
			sb.WriteString(fmt.Sprintf("\n## 动态（共 %d 篇）\n", len(authorPosts.Motions)))
			for _, post := range authorPosts.Motions {
				sb.WriteString(formatPostLine(post))
			}
		}

		// 输出作品集（按名称排序以保证稳定输出）
		albumNames := make([]string, 0, len(authorPosts.Albums))
		for name := range authorPosts.Albums {
			albumNames = append(albumNames, name)
		}
		sort.Strings(albumNames)

		for _, name := range albumNames {
			posts := authorPosts.Albums[name]
			if len(posts) > 0 {
				sb.WriteString(fmt.Sprintf("\n## 作品集：%s（共 %d 篇）\n", name, len(posts)))
				for _, post := range posts {
					sb.WriteString(formatPostLine(post))
				}
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}

// handleReadPost 处理 read_post Tool 调用
// 支持通过路径或作者+标题关键词读取文章内容
func handleReadPost(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := request.GetString("path", "")
		author := request.GetString("author", "")
		title := request.GetString("title", "")

		// 模式一：通过路径直接读取
		if path != "" {
			content, err := storage.ReadPost(dataDir, path)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("📄 文件：%s\n\n%s", path, content)), nil
		}

		// 模式二：通过作者+标题关键词匹配
		if author != "" && title != "" {
			matches, err := storage.FindPostByTitle(dataDir, author, title)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			if len(matches) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("未找到标题包含 '%s' 的文章", title)), nil
			}

			// 唯一匹配：直接返回内容
			if len(matches) == 1 {
				content, err := storage.ReadPost(dataDir, matches[0].Path)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				return mcp.NewToolResultText(fmt.Sprintf("📄 文件：%s\n\n%s", matches[0].Path, content)), nil
			}

			// 多个匹配：返回列表供用户选择
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("标题包含 '%s' 的文章有 %d 篇，请指定具体路径：\n\n", title, len(matches)))
			for _, m := range matches {
				sb.WriteString(formatPostLine(m))
			}
			return mcp.NewToolResultText(sb.String()), nil
		}

		return mcp.NewToolResultError("请提供 path 或 author+title 参数"), nil
	}
}

// handleSearch 处理 search Tool 调用
// 在已下载文档中全文搜索关键词
func handleSearch(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("请提供搜索关键词"), nil
		}

		author := request.GetString("author", "")

		resp, err := Search(dataDir, query, author, 20)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(formatSearchResponse(resp)), nil
	}
}

// formatPostLine 格式化单篇文章的输出行
func formatPostLine(post storage.PostInfo) string {
	if post.PublishTime != "" {
		return fmt.Sprintf("- [%s] %s → %s\n", post.PublishTime, post.Title, post.Path)
	}
	return fmt.Sprintf("- %s → %s\n", post.Title, post.Path)
}
