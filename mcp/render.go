package mcp

import (
	"AfdianToMarkdown/storage"
	"fmt"
	"strings"
)

// ---- JSON DTO（structuredContent 的线上结构） ----

type authorStat struct {
	Author    string `json:"author"`
	PostCount int    `json:"post_count"`
}

type authorsOut struct {
	Total   int          `json:"total"`
	Authors []authorStat `json:"authors"`
}

type postDTO struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Date          string   `json:"date"`
	Collections   []string `json:"collections"`
	CanonicalPath string   `json:"canonical_path"`
	URL           string   `json:"url"`
	WordCount     int      `json:"word_count"`
	HasComments   bool     `json:"has_comments"`
}

type postsOut struct {
	Author     string    `json:"author"`
	TotalCount int       `json:"total_count"`
	Returned   int       `json:"returned"`
	Offset     int       `json:"offset"`
	NextOffset *int      `json:"next_offset"`
	Posts      []postDTO `json:"posts"`
}

type readPostOut struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Date          string   `json:"date"`
	Collections   []string `json:"collections"`
	CanonicalPath string   `json:"canonical_path"`
	URL           string   `json:"url"`
	WordCount     int      `json:"word_count"`
	HasComments   bool     `json:"has_comments"`
	Content       string   `json:"content"`
}

type candidateDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Date  string `json:"date"`
	Path  string `json:"path"`
}

type disambiguationOut struct {
	Message    string         `json:"message"`
	Candidates []candidateDTO `json:"candidates"`
}

type snippetDTO struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

type searchDocDTO struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Author      string       `json:"author"`
	Date        string       `json:"date"`
	Path        string       `json:"path"`
	URL         string       `json:"url"`
	Collections []string     `json:"collections"`
	Score       int          `json:"score"`
	HitCount    int          `json:"hit_count"`
	Snippets    []snippetDTO `json:"snippets"`
}

type searchOut struct {
	Query        string         `json:"query"`
	Match        string         `json:"match"`
	TotalDocs    int            `json:"total_docs"`
	TotalHits    int            `json:"total_hits"`
	ReturnedDocs int            `json:"returned_docs"`
	Offset       int            `json:"offset"`
	NextOffset   *int           `json:"next_offset"`
	Results      []searchDocDTO `json:"results"`
}

type commentDTO struct {
	Index     int    `json:"index"`
	Time      string `json:"time"`
	Commenter string `json:"commenter"`
	Text      string `json:"text"`
	IsHot     bool   `json:"is_hot"`
}

type commentsOut struct {
	ID         string       `json:"id"`
	Title      string       `json:"title"`
	Path       string       `json:"path"`
	TotalCount int          `json:"total_count"`
	Returned   int          `json:"returned"`
	Offset     int          `json:"offset"`
	NextOffset *int         `json:"next_offset"`
	Comments   []commentDTO `json:"comments"`
}

// ---- 辅助 ----

// nextOffset 计算下一页偏移，无更多数据时返回 nil
func nextOffset(offset, returned, total int) *int {
	n := offset + returned
	if n < total {
		return &n
	}
	return nil
}

// paginate 返回切片在 [offset, offset+limit) 的边界（已做越界裁剪）
func paginate(total, offset, limit int) (start, end int) {
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end = offset + limit
	if limit <= 0 || end > total {
		end = total
	}
	return offset, end
}

func toPostDTO(p storage.Post) postDTO {
	return postDTO{
		ID:            p.ID,
		Title:         p.Title,
		Date:          p.Date,
		Collections:   p.Collections,
		CanonicalPath: p.CanonicalPath,
		URL:           p.URL,
		WordCount:     p.WordCount,
		HasComments:   p.HasComments,
	}
}

func collectionsTag(cols []string) string {
	if len(cols) == 0 {
		return ""
	}
	return " [" + strings.Join(cols, ", ") + "]"
}

// ---- 文本镜像渲染 ----

func renderAuthorsText(out authorsOut) string {
	if out.Total == 0 {
		return "当前没有已下载的作者。"
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "已下载作者（共 %d 位）：\n", out.Total)
	for _, a := range out.Authors {
		fmt.Fprintf(&sb, "- %s（%d 篇）\n", a.Author, a.PostCount)
	}
	return sb.String()
}

func renderPostsText(out postsOut) string {
	if out.TotalCount == 0 {
		return fmt.Sprintf("作者 %s 没有符合条件的文章。", out.Author)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "作者 %s：共 %d 篇（显示 %d-%d）\n",
		out.Author, out.TotalCount, out.Offset+1, out.Offset+out.Returned)
	for _, p := range out.Posts {
		date := p.Date
		if date == "" {
			date = "无日期"
		}
		fmt.Fprintf(&sb, "- [%s] %s（%d字%s）%s → %s\n",
			date, p.Title, p.WordCount, commentMark(p.HasComments), collectionsTag(p.Collections), p.CanonicalPath)
	}
	return sb.String()
}

func commentMark(has bool) string {
	if has {
		return "，有评论"
	}
	return ""
}

func renderReadPostText(out readPostOut) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "📄 %s\n", out.CanonicalPath)
	fmt.Fprintf(&sb, "[%s] %s", out.Date, out.Title)
	if out.URL != "" {
		fmt.Fprintf(&sb, "  •  %s", out.URL)
	}
	sb.WriteString("\n\n")
	sb.WriteString(out.Content)
	return sb.String()
}

func renderDisambiguationText(out disambiguationOut) string {
	var sb strings.Builder
	sb.WriteString(out.Message + "\n")
	for _, c := range out.Candidates {
		fmt.Fprintf(&sb, "- [%s] %s  (id: %s)  → %s\n", c.Date, c.Title, c.ID, c.Path)
	}
	return sb.String()
}

func renderSearchText(out searchOut) string {
	if out.TotalDocs == 0 {
		return fmt.Sprintf("未找到包含 %q 的内容。", out.Query)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "搜索 %q（match=%s）：%d 篇 / %d 次命中（显示 %d-%d）\n",
		out.Query, out.Match, out.TotalDocs, out.TotalHits, out.Offset+1, out.Offset+out.ReturnedDocs)
	for _, r := range out.Results {
		fmt.Fprintf(&sb, "\n--- [%s] %s（%d 次）%s → %s\n", r.Date, r.Title, r.HitCount, collectionsTag(r.Collections), r.Path)
		for _, s := range r.Snippets {
			sb.WriteString(s.Text + "\n")
		}
	}
	if out.NextOffset != nil {
		fmt.Fprintf(&sb, "\n更多结果：offset=%d\n", *out.NextOffset)
	}
	return sb.String()
}

func renderCommentsText(out commentsOut) string {
	if out.TotalCount == 0 {
		return fmt.Sprintf("《%s》没有符合条件的评论。", out.Title)
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "《%s》（%s）：共 %d 条评论（显示 %d-%d）\n",
		out.Title, out.Path, out.TotalCount, out.Offset+1, out.Offset+out.Returned)
	for _, c := range out.Comments {
		tag := "评论"
		if c.IsHot {
			tag = "热评"
		}
		fmt.Fprintf(&sb, "\n[%s] %s by %s\n%s\n", tag, c.Time, c.Commenter, c.Text)
	}
	return sb.String()
}
