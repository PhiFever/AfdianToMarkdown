package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostInfo 表示一篇文章的元信息（不含内容），用于从文件名解析标题与日期
type PostInfo struct {
	Title       string // 文章标题（从文件名中提取，去除时间戳前缀）
	Path        string // 相对于数据目录的文件路径
	Category    string // 类别："motions" 或作品集名称
	PublishTime string // 发布时间（从文件名提取，格式 YYYY-MM-DD）
}

func errAuthorNotExist(author string) error {
	return fmt.Errorf("作者不存在：%s", author)
}

// safePath 验证相对路径解析后仍在 dataDir 内，防止路径遍历攻击。
// 返回规范化后的绝对路径，如果路径逃逸则返回错误。
func safePath(dataDir, relativePath string) (string, error) {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(absDataDir, relativePath)
	cleanPath := filepath.Clean(fullPath)
	// 确保解析后的路径仍在数据目录内（加 separator 防止前缀误匹配，如 /data2 匹配 /data）
	if !strings.HasPrefix(cleanPath, absDataDir+string(filepath.Separator)) && cleanPath != absDataDir {
		return "", fmt.Errorf("路径不合法：%s", relativePath)
	}
	return cleanPath, nil
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

// ReadPost 读取指定相对路径的 Markdown 文件，返回完整内容
func ReadPost(dataDir, relativePath string) (string, error) {
	// 安全检查：防止路径遍历
	fullPath, err := safePath(dataDir, relativePath)
	if err != nil {
		return "", fmt.Errorf("文件不存在：%s", relativePath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("文件不存在：%s", relativePath)
		}
		return "", err
	}
	return string(data), nil
}
