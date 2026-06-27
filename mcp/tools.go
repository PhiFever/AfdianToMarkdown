package mcp

import (
	"AfdianToMarkdown/storage"
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleListAuthors 处理 list_authors：返回每位作者及其去重后的文章数
func handleListAuthors(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authors, err := storage.ListAuthors(dataDir)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		stats := make([]authorStat, 0, len(authors))
		for _, a := range authors {
			posts, err := storage.BuildAuthorIndex(dataDir, a)
			if err != nil {
				continue
			}
			stats = append(stats, authorStat{Author: a, PostCount: len(posts)})
		}

		out := authorsOut{Total: len(stats), Authors: stats}
		return mcp.NewToolResultStructured(out, renderAuthorsText(out)), nil
	}
}

// handleListPosts 处理 list_posts：过滤 + 分页 + 精简投影（按 canonical id 去重）
func handleListPosts(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		author, err := request.RequireString("author")
		if err != nil {
			return mcp.NewToolResultError("请提供 author 参数"), nil
		}
		collection := request.GetString("collection", "")
		titleContains := strings.ToLower(request.GetString("title_contains", ""))
		dateFrom := request.GetString("date_from", "")
		dateTo := request.GetString("date_to", "")
		limit := request.GetInt("limit", 50)
		offset := request.GetInt("offset", 0)

		posts, err := storage.BuildAuthorIndex(dataDir, author)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		filtered := make([]storage.Post, 0, len(posts))
		for _, p := range posts {
			if collection != "" && !storage.HasCollection(p, collection) {
				continue
			}
			if titleContains != "" && !strings.Contains(strings.ToLower(p.Title), titleContains) {
				continue
			}
			if !dateInRange(p.Date, dateFrom, dateTo) {
				continue
			}
			filtered = append(filtered, p)
		}

		total := len(filtered)
		start, end := paginate(total, offset, limit)
		page := filtered[start:end]

		dtos := make([]postDTO, 0, len(page))
		for _, p := range page {
			dtos = append(dtos, toPostDTO(p))
		}

		out := postsOut{
			Author:     author,
			TotalCount: total,
			Returned:   len(dtos),
			Offset:     start,
			NextOffset: nextOffset(start, len(dtos), total),
			Posts:      dtos,
		}
		return mcp.NewToolResultStructured(out, renderPostsText(out)), nil
	}
}

// handleReadPost 处理 read_post：id → url → path → author+title
func handleReadPost(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := request.GetString("id", "")
		url := request.GetString("url", "")
		path := request.GetString("path", "")
		author := request.GetString("author", "")
		title := request.GetString("title", "")

		// 模式一/二/三：id / url / path
		if id != "" || url != "" || path != "" {
			post, content, err := resolvePost(dataDir, id, url, path)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			out := toReadPostOut(post, content)
			return mcp.NewToolResultStructured(out, renderReadPostText(out)), nil
		}

		// 模式四：作者 + 标题（按 id 去重）
		if author != "" && title != "" {
			posts, err := storage.BuildAuthorIndex(dataDir, author)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			matches := storage.FindByTitle(posts, title)
			switch len(matches) {
			case 0:
				return mcp.NewToolResultError(fmt.Sprintf("未找到标题包含 '%s' 的文章", title)), nil
			case 1:
				content, err := storage.ReadPost(dataDir, matches[0].CanonicalPath)
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				out := toReadPostOut(matches[0], content)
				return mcp.NewToolResultStructured(out, renderReadPostText(out)), nil
			default:
				out := disambiguationOut{
					Message:    fmt.Sprintf("标题包含 '%s' 的文章有 %d 篇，请用 id 或 path 指定：", title, len(matches)),
					Candidates: toCandidates(matches),
				}
				return mcp.NewToolResultStructured(out, renderDisambiguationText(out)), nil
			}
		}

		return mcp.NewToolResultError("请提供 id、url、path 或 author+title 参数"), nil
	}
}

// handleSearch 处理 search：分词 + 文档级聚合 + 启发式排序 + 分页
func handleSearch(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError("请提供搜索关键词"), nil
		}

		p := searchParams{
			query:      query,
			matchAll:   request.GetString("match", "all") != "any",
			author:     request.GetString("author", ""),
			collection: request.GetString("collection", ""),
			limit:      request.GetInt("limit", 10),
			offset:     request.GetInt("offset", 0),
			perDocHits: request.GetInt("per_doc_hits", 3),
		}

		out, err := Search(dataDir, p)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultStructured(*out, renderSearchText(*out)), nil
	}
}

// handleListComments 处理 list_comments：定位文章 + 解析评论 + 过滤 + 分页
func handleListComments(dataDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := request.GetString("id", "")
		url := request.GetString("url", "")
		path := request.GetString("path", "")
		if id == "" && url == "" && path == "" {
			return mcp.NewToolResultError("请提供 id、url 或 path 参数"), nil
		}
		commenter := strings.ToLower(request.GetString("commenter", ""))
		hotOnly := request.GetBool("hot_only", false)
		limit := request.GetInt("limit", 50)
		offset := request.GetInt("offset", 0)

		post, content, err := resolvePost(dataDir, id, url, path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		all := storage.ParseComments(content)
		filtered := make([]storage.Comment, 0, len(all))
		for _, c := range all {
			if hotOnly && !c.IsHot {
				continue
			}
			if commenter != "" && !strings.Contains(strings.ToLower(c.Commenter), commenter) {
				continue
			}
			filtered = append(filtered, c)
		}

		total := len(filtered)
		start, end := paginate(total, offset, limit)
		page := filtered[start:end]

		dtos := make([]commentDTO, 0, len(page))
		for _, c := range page {
			dtos = append(dtos, commentDTO{
				Index:     c.Index,
				Time:      c.Time,
				Commenter: c.Commenter,
				Text:      c.Text,
				IsHot:     c.IsHot,
			})
		}

		out := commentsOut{
			ID:         post.ID,
			Title:      post.Title,
			Path:       post.CanonicalPath,
			TotalCount: total,
			Returned:   len(dtos),
			Offset:     start,
			NextOffset: nextOffset(start, len(dtos), total),
			Comments:   dtos,
		}
		return mcp.NewToolResultStructured(out, renderCommentsText(out)), nil
	}
}

// resolvePost 按 id / url / path 定位一篇文章，返回 Post 与完整内容
func resolvePost(dataDir, id, url, path string) (storage.Post, string, error) {
	switch {
	case id != "" || url != "":
		hash := storage.ExtractHash(id)
		if url != "" {
			hash = storage.ExtractHash(url)
		}
		posts, err := storage.BuildIndex(dataDir)
		if err != nil {
			return storage.Post{}, "", err
		}
		post, ok := storage.FindByID(posts, hash)
		if !ok {
			return storage.Post{}, "", fmt.Errorf("未找到 id 为 '%s' 的文章", hash)
		}
		content, err := storage.ReadPost(dataDir, post.CanonicalPath)
		if err != nil {
			return storage.Post{}, "", err
		}
		return post, content, nil

	default: // path
		content, err := storage.ReadPost(dataDir, path)
		if err != nil {
			return storage.Post{}, "", err
		}
		// 尝试用作者索引补全元信息
		segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(segments) > 0 {
			if posts, err := storage.BuildAuthorIndex(dataDir, segments[0]); err == nil {
				for _, p := range posts {
					if p.CanonicalPath == path {
						return p, content, nil
					}
				}
			}
		}
		// 回退：从文件名与内容估算元信息（如传入的是非 canonical 副本路径）
		base := segments[len(segments)-1]
		meta := storage.ParsePostInfo(base, "", "")
		wc, hc := storage.AnalyzeContent(content)
		author := ""
		if len(segments) > 0 {
			author = segments[0]
		}
		return storage.Post{
			Title:         meta.Title,
			Author:        author,
			Date:          meta.PublishTime,
			Collections:   []string{},
			CanonicalPath: path,
			WordCount:     wc,
			HasComments:   hc,
		}, content, nil
	}
}

func toReadPostOut(p storage.Post, content string) readPostOut {
	return readPostOut{
		ID:            p.ID,
		Title:         p.Title,
		Author:        p.Author,
		Date:          p.Date,
		Collections:   p.Collections,
		CanonicalPath: p.CanonicalPath,
		URL:           p.URL,
		WordCount:     p.WordCount,
		HasComments:   p.HasComments,
		Content:       content,
	}
}

func toCandidates(posts []storage.Post) []candidateDTO {
	out := make([]candidateDTO, 0, len(posts))
	for _, p := range posts {
		out = append(out, candidateDTO{
			ID:    p.ID,
			Title: p.Title,
			Date:  p.Date,
			Path:  p.CanonicalPath,
		})
	}
	return out
}

// dateInRange 判断日期是否落在 [from, to] 内（YYYY-MM-DD 字典序比较，空界限不限制）
func dateInRange(date, from, to string) bool {
	if from == "" && to == "" {
		return true
	}
	if date == "" {
		return false // 设了日期过滤但文章无日期 → 排除
	}
	if from != "" && date < from {
		return false
	}
	if to != "" && date > to {
		return false
	}
	return true
}
