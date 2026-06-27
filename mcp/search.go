package mcp

import (
	"AfdianToMarkdown/storage"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

// searchParams 聚合一次搜索的全部参数
type searchParams struct {
	query      string
	matchAll   bool // true=AND，false=OR
	author     string
	collection string
	limit      int
	offset     int
	perDocHits int
}

// parseQuery 将查询串解析为词项：
//   - `"..."` 段为短语，精确子串匹配
//   - 其余按空白/标点切分为词
//
// 全部转小写并去重。不做中文分词，无空格的中文串将作为单一词项。
func parseQuery(query string) []string {
	var terms []string
	var sb strings.Builder
	inQuote := false

	flushPlain := func() {
		fields := strings.FieldsFunc(sb.String(), func(r rune) bool {
			return unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r)
		})
		sb.Reset()
		for _, f := range fields {
			terms = append(terms, strings.ToLower(f))
		}
	}

	for _, r := range query {
		if r == '"' {
			if inQuote {
				if p := strings.TrimSpace(sb.String()); p != "" {
					terms = append(terms, strings.ToLower(p))
				}
				sb.Reset()
				inQuote = false
			} else {
				flushPlain()
				inQuote = true
			}
			continue
		}
		sb.WriteRune(r)
	}
	if inQuote {
		// 未闭合引号：剩余部分作为短语处理
		if p := strings.TrimSpace(sb.String()); p != "" {
			terms = append(terms, strings.ToLower(p))
		}
	} else {
		flushPlain()
	}

	return dedupStrings(terms)
}

func dedupStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// docMatch 单篇文档的命中聚合
type docMatch struct {
	post       storage.Post
	hitCount   int   // 命中行数
	coverage   int   // 命中的不同词项数
	titleMatch bool  // 标题是否命中
	score      int   // 信息性相关度分值
	matchLines []int // 命中行号（0-based）
	lines      []string
}

// Search 执行全文搜索：按 canonical id 去重、文档级聚合、启发式排序、分页
func Search(dataDir string, p searchParams) (*searchOut, error) {
	terms := parseQuery(p.query)
	if len(terms) == 0 {
		return nil, fmt.Errorf("请提供有效的搜索关键词")
	}

	var posts []storage.Post
	var err error
	if p.author != "" {
		posts, err = storage.BuildAuthorIndex(dataDir, p.author)
	} else {
		posts, err = storage.BuildIndex(dataDir)
	}
	if err != nil {
		return nil, err
	}

	var matches []docMatch
	totalHits := 0
	for _, post := range posts {
		if p.collection != "" && !storage.HasCollection(post, p.collection) {
			continue
		}
		content, err := storage.ReadPost(dataDir, post.CanonicalPath)
		if err != nil {
			continue
		}
		m, ok := matchDocument(post, content, terms, p.matchAll)
		if !ok {
			continue
		}
		totalHits += m.hitCount
		matches = append(matches, m)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if a.coverage != b.coverage {
			return a.coverage > b.coverage
		}
		if a.titleMatch != b.titleMatch {
			return a.titleMatch
		}
		if a.hitCount != b.hitCount {
			return a.hitCount > b.hitCount
		}
		if a.post.Date != b.post.Date {
			return a.post.Date > b.post.Date
		}
		return a.post.CanonicalPath < b.post.CanonicalPath
	})

	totalDocs := len(matches)
	start, end := paginate(totalDocs, p.offset, p.limit)
	page := matches[start:end]

	results := make([]searchDocDTO, 0, len(page))
	for _, m := range page {
		results = append(results, searchDocDTO{
			ID:          m.post.ID,
			Title:       m.post.Title,
			Author:      m.post.Author,
			Date:        m.post.Date,
			Path:        m.post.CanonicalPath,
			URL:         m.post.URL,
			Collections: m.post.Collections,
			Score:       m.score,
			HitCount:    m.hitCount,
			Snippets:    buildSnippets(m, p.perDocHits),
		})
	}

	return &searchOut{
		Query:        p.query,
		Match:        matchMode(p.matchAll),
		TotalDocs:    totalDocs,
		TotalHits:    totalHits,
		ReturnedDocs: len(results),
		Offset:       start,
		NextOffset:   nextOffset(start, len(results), totalDocs),
		Results:      results,
	}, nil
}

// matchDocument 在单篇内容中匹配全部词项，返回聚合命中信息
func matchDocument(post storage.Post, content string, terms []string, matchAll bool) (docMatch, bool) {
	lines := strings.Split(content, "\n")

	covered := make(map[string]struct{}, len(terms))
	var matchLines []int
	for i, line := range lines {
		ll := strings.ToLower(line)
		hit := false
		for _, t := range terms {
			if strings.Contains(ll, t) {
				covered[t] = struct{}{}
				hit = true
			}
		}
		if hit {
			matchLines = append(matchLines, i)
		}
	}

	titleLower := strings.ToLower(post.Title)
	titleMatch := false
	for _, t := range terms {
		if strings.Contains(titleLower, t) {
			covered[t] = struct{}{}
			titleMatch = true
		}
	}

	coverage := len(covered)
	if coverage == 0 || (matchAll && coverage < len(terms)) {
		return docMatch{}, false
	}

	score := coverage*1_000_000 + len(matchLines)
	if titleMatch {
		score += 500_000
	}

	return docMatch{
		post:       post,
		hitCount:   len(matchLines),
		coverage:   coverage,
		titleMatch: titleMatch,
		score:      score,
		matchLines: matchLines,
		lines:      lines,
	}, true
}

// buildSnippets 取前 perDocHits 条命中行，附带上下文
func buildSnippets(m docMatch, perDocHits int) []snippetDTO {
	if perDocHits <= 0 {
		perDocHits = 3
	}
	snippets := make([]snippetDTO, 0, perDocHits)
	for _, idx := range m.matchLines {
		if len(snippets) >= perDocHits {
			break
		}
		snippets = append(snippets, snippetDTO{
			Line: idx + 1,
			Text: buildContext(m.lines, idx),
		})
	}
	return snippets
}

// buildContext 构建匹配行及前后各 2 行的上下文文本
func buildContext(lines []string, matchIndex int) string {
	start := matchIndex - 2
	if start < 0 {
		start = 0
	}
	end := matchIndex + 2
	if end > len(lines)-1 {
		end = len(lines) - 1
	}
	var sb strings.Builder
	for i := start; i <= end; i++ {
		marker := "  "
		if i == matchIndex {
			marker = "> "
		}
		fmt.Fprintf(&sb, "%s%d | %s\n", marker, i+1, lines[i])
	}
	return strings.TrimRight(sb.String(), "\n")
}

func matchMode(all bool) string {
	if all {
		return "all"
	}
	return "any"
}
