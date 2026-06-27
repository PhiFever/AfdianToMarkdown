package storage

import (
	"regexp"
	"strconv"
	"strings"
)

// Comment 表示一条评论
type Comment struct {
	Index     int    // 段内序号 [N]
	Time      string // YYYY-MM-DD HH:MM:SS
	Commenter string // 评论者名
	Text      string // 评论正文
	IsHot     bool   // true=热评(## 热评)，false=普通(## 评论)
}

// commentHeaderRegexp 匹配形如 ##### <span>[0] 2020-12-15 20:30:16 by ciciff</span>
var commentHeaderRegexp = regexp.MustCompile(`^#{5}\s*<span>\[(\d+)\]\s+(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\s+by\s+(.+?)</span>`)

// ParseComments 解析 Markdown 中的 ## 热评 / ## 评论 区，返回结构化评论列表
func ParseComments(content string) []Comment {
	lines := strings.Split(content, "\n")
	comments := make([]Comment, 0)

	isHot := false
	inSection := false
	var cur *Comment
	var textLines []string

	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(strings.Join(textLines, "\n"))
			comments = append(comments, *cur)
			cur = nil
			textLines = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 段落切换：## 开头的二级标题
		if strings.HasPrefix(trimmed, "## ") {
			flush()
			switch trimmed {
			case "## 热评":
				isHot = true
				inSection = true
			case "## 评论":
				isHot = false
				inSection = true
			default:
				inSection = false
			}
			continue
		}
		if !inSection {
			continue
		}

		if m := commentHeaderRegexp.FindStringSubmatch(trimmed); m != nil {
			flush()
			idx, _ := strconv.Atoi(m[1])
			cur = &Comment{Index: idx, Time: m[2], Commenter: m[3], IsHot: isHot}
			continue
		}
		if trimmed == "----" {
			continue // 条目分隔符
		}
		if cur != nil {
			textLines = append(textLines, line)
		}
	}
	flush()
	return comments
}
