package storage

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Post 表示去重后的一篇文章（按 canonical id 聚合 motions 与各作品集中的多份副本）
type Post struct {
	ID            string   // canonical id = ### Refer 中 afdian.com/p/<hash> 的 hash
	Title         string   // 文件名解析（去时间戳前缀）
	Author        string   // 作者目录名
	Date          string   // YYYY-MM-DD（文件名解析）
	Collections   []string // 该 id 所属作品集名（不含 motions），排序去重
	CanonicalPath string   // 去重代表路径：优先 motions/，否则字母序首个 album
	URL           string   // https://afdian.com/p/<hash>
	WordCount     int      // 正文 rune 数（剔除图片/媒体/评论），估算
	HasComments   bool     // 是否含 ## 热评 或 ## 评论
}

// referRegexp 提取 Refer 中的 post hash，兼容两种形态：
//   - afdian.com/p/<postHash>                （动态下载）
//   - afdian.com/album/<albumId>/<postHash>  （作品集下载）
//
// 两种形态下同一篇文章的 postHash 一致，作为 canonical id。
var referRegexp = regexp.MustCompile(`afdian\.com/(?:p|album/[a-zA-Z0-9]+)/([a-zA-Z0-9]+)`)

const motionsCategory = "motions"

// BuildAuthorIndex 扫描单个作者目录，返回按 canonical id 去重的文章列表（按日期倒序）
func BuildAuthorIndex(dataDir, author string) ([]Post, error) {
	authorDir, err := safePath(dataDir, author)
	if err != nil {
		return nil, errAuthorNotExist(author)
	}
	info, err := os.Stat(authorDir)
	if err != nil || !info.IsDir() {
		return nil, errAuthorNotExist(author)
	}

	entries, err := os.ReadDir(authorDir)
	if err != nil {
		return nil, err
	}

	// 按 hash 聚合同一篇文章在不同子目录下的副本
	type variant struct {
		category string
		relPath  string
		fileName string
	}
	groups := make(map[string][]variant)
	order := []string{} // 保持首次出现顺序，便于 hash 缺失时的稳定回退

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".assets" {
			continue
		}
		category := entry.Name()
		subDir := filepath.Join(authorDir, category)
		files, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
				continue
			}
			relPath := filepath.ToSlash(filepath.Join(author, category, f.Name()))
			hash := extractRefHash(filepath.Join(subDir, f.Name()))
			key := hash
			if key == "" {
				// 无 Refer hash：按路径自成一篇，避免跨文件误去重
				key = "path:" + relPath
			}
			if _, ok := groups[key]; !ok {
				order = append(order, key)
			}
			groups[key] = append(groups[key], variant{category: category, relPath: relPath, fileName: f.Name()})
		}
	}

	posts := make([]Post, 0, len(groups))
	for _, key := range order {
		vs := groups[key]

		// 选 canonical：优先 motions，否则字母序首个作品集
		sort.Slice(vs, func(i, j int) bool {
			if (vs[i].category == motionsCategory) != (vs[j].category == motionsCategory) {
				return vs[i].category == motionsCategory
			}
			if vs[i].category != vs[j].category {
				return vs[i].category < vs[j].category
			}
			return vs[i].relPath < vs[j].relPath
		})
		canonical := vs[0]

		// collections = 非 motions 的子目录名，去重排序
		collSet := make(map[string]struct{})
		for _, v := range vs {
			if v.category != motionsCategory {
				collSet[v.category] = struct{}{}
			}
		}
		collections := make([]string, 0, len(collSet))
		for c := range collSet {
			collections = append(collections, c)
		}
		sort.Strings(collections)

		meta := ParsePostInfo(canonical.fileName, canonical.category, "")
		hash := key
		url := ""
		if strings.HasPrefix(key, "path:") {
			hash = ""
		} else {
			url = "https://afdian.com/p/" + key
		}

		full, _ := os.ReadFile(filepathJoinData(dataDir, canonical.relPath))
		wordCount, hasComments := AnalyzeContent(string(full))

		posts = append(posts, Post{
			ID:            hash,
			Title:         meta.Title,
			Author:        author,
			Date:          meta.PublishTime,
			Collections:   collections,
			CanonicalPath: canonical.relPath,
			URL:           url,
			WordCount:     wordCount,
			HasComments:   hasComments,
		})
	}

	sortPostsByDateDesc(posts)
	return posts, nil
}

// BuildIndex 扫描所有作者，返回全量去重文章列表（按日期倒序）
func BuildIndex(dataDir string) ([]Post, error) {
	authors, err := ListAuthors(dataDir)
	if err != nil {
		return nil, err
	}
	var all []Post
	for _, a := range authors {
		posts, err := BuildAuthorIndex(dataDir, a)
		if err != nil {
			continue
		}
		all = append(all, posts...)
	}
	sortPostsByDateDesc(all)
	return all, nil
}

// extractRefHash 读取文件头部数行，提取 ### Refer 中的 afdian hash，缺失返回 ""
func extractRefHash(fullPath string) string {
	f, err := os.Open(fullPath)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for i := 0; i < 15 && scanner.Scan(); i++ {
		if m := referRegexp.FindStringSubmatch(scanner.Text()); m != nil {
			return m[1]
		}
	}
	return ""
}

// commentEntryMarker 标识一条真实评论条目（写入器总会输出评论区标题，
// 故以是否存在条目行来判断「确有评论」）
const commentEntryMarker = "##### <span>["

// AnalyzeContent 估算正文字数（rune 计数，剔除图片/媒体/评论）并判断是否含真实评论
func AnalyzeContent(content string) (wordCount int, hasComments bool) {
	hasComments = strings.Contains(content, commentEntryMarker)

	// 评论区之前的部分视为正文范围
	body := content
	if idx := firstCommentIndex(content); idx >= 0 {
		body = content[:idx]
	}

	count := 0
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue // 跳过空行与标题/Refer 等
		}
		if strings.HasPrefix(trimmed, "![") || strings.HasPrefix(trimmed, "<audio") || strings.HasPrefix(trimmed, "<video") {
			continue // 跳过图片与媒体标签
		}
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			continue // 跳过 Refer URL 等纯链接行
		}
		for _, r := range trimmed {
			if !isSpace(r) {
				count++
			}
		}
	}
	return count, hasComments
}

func firstCommentIndex(content string) int {
	hot := strings.Index(content, "## 热评")
	normal := strings.Index(content, "## 评论")
	switch {
	case hot < 0:
		return normal
	case normal < 0:
		return hot
	case hot < normal:
		return hot
	default:
		return normal
	}
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '　'
}

func sortPostsByDateDesc(posts []Post) {
	sort.SliceStable(posts, func(i, j int) bool {
		if posts[i].Date != posts[j].Date {
			return posts[i].Date > posts[j].Date
		}
		return posts[i].Title < posts[j].Title
	})
}

// filepathJoinData 拼接数据目录与相对路径（相对路径已用 / 分隔）
func filepathJoinData(dataDir, relPath string) string {
	return filepath.Join(dataDir, filepath.FromSlash(relPath))
}

// FindByID 在去重列表中按 canonical id 精确查找
func FindByID(posts []Post, id string) (Post, bool) {
	for _, p := range posts {
		if p.ID == id {
			return p, true
		}
	}
	return Post{}, false
}

// FindByTitle 按标题关键词子串匹配（大小写不敏感）
func FindByTitle(posts []Post, titleKeyword string) []Post {
	keyword := strings.ToLower(titleKeyword)
	var matches []Post
	for _, p := range posts {
		if strings.Contains(strings.ToLower(p.Title), keyword) {
			matches = append(matches, p)
		}
	}
	return matches
}

// ExtractHash 从 afdian URL 或裸 hash 中提取 canonical id
func ExtractHash(urlOrHash string) string {
	if m := referRegexp.FindStringSubmatch(urlOrHash); m != nil {
		return m[1]
	}
	return strings.TrimSpace(urlOrHash)
}

// HasCollection 判断文章是否属于指定作品集
func HasCollection(p Post, collection string) bool {
	for _, c := range p.Collections {
		if c == collection {
			return true
		}
	}
	return false
}
