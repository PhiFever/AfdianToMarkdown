package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostInfo 表示一篇文章的元信息（不含内容）
type PostInfo struct {
	Title       string // 文章标题（从文件名中提取，去除时间戳前缀）
	Path        string // 相对于数据目录的文件路径
	Category    string // 类别："motions" 或作品集名称
	PublishTime string // 发布时间（从文件名提取，格式 YYYY-MM-DD）
}

// AuthorPosts 表示一位作者下所有文章的分组结构
type AuthorPosts struct {
	Author  string                // 作者名（目录名）
	Motions []PostInfo            // 动态列表
	Albums  map[string][]PostInfo // 作品集名 → 文章列表
}

// SearchResult 表示一条搜索匹配结果
type SearchResult struct {
	FilePath   string // 相对于数据目录的文件路径
	Title      string // 文章标题
	Author     string // 作者名
	LineNumber int    // 匹配行号
	Context    string // 匹配行及前后各 3 行的上下文文本
}

// SearchResponse 表示搜索的完整返回
type SearchResponse struct {
	Query      string         // 搜索关键词
	TotalCount int            // 总匹配数
	Results    []SearchResult // 返回的匹配结果（上限 20 条）
	Truncated  bool           // 是否因超过上限而截断
}

// ListAuthors 扫描数据目录，返回所有作者名称列表
func ListAuthors(dataDir string) ([]string, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("数据目录不存在：%s", dataDir)
		}
		return nil, err
	}

	var authors []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".assets" {
			authors = append(authors, entry.Name())
		}
	}
	return authors, nil
}

// ParsePostInfo 从文件名中解析文章元信息
// 文件名格式：{YYYY-MM-DD_HH_MM_SS}_{SafeTitle}.md
// 前 19 个字符为时间戳，第 20 个字符为分隔符 _，第 21 个字符起为安全标题
func ParsePostInfo(fileName, category, relativeDir string) PostInfo {
	title := strings.TrimSuffix(fileName, ".md")
	publishTime := ""

	// 尝试从文件名中提取时间戳和标题
	if len(fileName) > 20 && fileName[19] == '_' {
		publishTime = fileName[:10] // YYYY-MM-DD
		title = strings.TrimSuffix(fileName[20:], ".md")
	}

	return PostInfo{
		Title:       title,
		Path:        filepath.ToSlash(filepath.Join(relativeDir, fileName)),
		Category:    category,
		PublishTime: publishTime,
	}
}

// ListPosts 扫描指定作者目录，返回按动态和作品集分组的文章列表
func ListPosts(dataDir, author string) (*AuthorPosts, error) {
	authorDir := filepath.Join(dataDir, author)
	info, err := os.Stat(authorDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("作者不存在：%s", author)
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("作者不存在：%s", author)
	}

	result := &AuthorPosts{
		Author: author,
		Albums: make(map[string][]PostInfo),
	}

	// 扫描作者目录下的子目录
	entries, err := os.ReadDir(authorDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".assets" {
			continue
		}

		subDirName := entry.Name()
		subDirPath := filepath.Join(authorDir, subDirName)
		relativeDir := filepath.ToSlash(filepath.Join(author, subDirName))

		// 扫描子目录中的 .md 文件
		files, err := os.ReadDir(subDirPath)
		if err != nil {
			continue
		}

		var posts []PostInfo
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".md") {
				continue
			}
			posts = append(posts, ParsePostInfo(f.Name(), subDirName, relativeDir))
		}

		if subDirName == "motions" {
			result.Motions = posts
		} else {
			result.Albums[subDirName] = posts
		}
	}

	return result, nil
}

// ReadPost 读取指定相对路径的 Markdown 文件，返回完整内容
func ReadPost(dataDir, relativePath string) (string, error) {
	// 安全检查：防止路径遍历
	cleanPath := filepath.Clean(relativePath)
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("文件不存在：%s", relativePath)
	}

	fullPath := filepath.Join(dataDir, cleanPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("文件不存在：%s", relativePath)
		}
		return "", err
	}
	return string(data), nil
}

// FindPostByTitle 按标题关键词模糊匹配，在指定作者的所有文章中搜索
// 返回匹配的文章列表（大小写不敏感）
func FindPostByTitle(dataDir, author, titleKeyword string) ([]PostInfo, error) {
	authorPosts, err := ListPosts(dataDir, author)
	if err != nil {
		return nil, err
	}

	keyword := strings.ToLower(titleKeyword)
	var matches []PostInfo

	// 搜索动态
	for _, post := range authorPosts.Motions {
		if strings.Contains(strings.ToLower(post.Title), keyword) {
			matches = append(matches, post)
		}
	}

	// 搜索所有作品集
	for _, posts := range authorPosts.Albums {
		for _, post := range posts {
			if strings.Contains(strings.ToLower(post.Title), keyword) {
				matches = append(matches, post)
			}
		}
	}

	return matches, nil
}
