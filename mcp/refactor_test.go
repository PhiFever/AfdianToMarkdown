package mcp

import (
	"AfdianToMarkdown/storage"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// ---- 测试夹具 ----

func mkPost(title, hash string, bodyLines []string, withComments bool) string {
	return mkPostRefer(title, "https://afdian.com/p/"+hash, bodyLines, withComments)
}

// mkAlbumPost 生成作品集形态 Refer（afdian.com/album/<albumId>/<postHash>）的副本
func mkAlbumPost(title, hash string, bodyLines []string, withComments bool) string {
	return mkPostRefer(title, "https://afdian.com/album/albumXYZ/"+hash, bodyLines, withComments)
}

func mkPostRefer(title, refer string, bodyLines []string, withComments bool) string {
	content := fmt.Sprintf("## %s\n\n### Refer\n\n%s\n\n### 正文\n\n%s\n",
		title, refer, strings.Join(bodyLines, "\n"))
	if withComments {
		content += "\n## 热评\n\n" +
			"----\n##### <span>[0] 2021-01-01 10:00:00 by 张三</span>\n\n热评内容一\n\n" +
			"----\n##### <span>[1] 2021-01-02 11:00:00 by 李四</span>\n\n热评内容二\n\n" +
			"## 评论\n\n" +
			"----\n##### <span>[0] 2022-01-01 10:00:00 by 王五</span>\n\n普通评论\n"
	}
	return content
}

func writeFixtureFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setupFixture 构造覆盖去重/回退/评论的样例数据目录
func setupFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// alice：含 motions 超集 + 同一篇在多个作品集中的副本
	writeFixtureFile(t, dir, "alice/motions/2022-03-02_10_00_00_时间表.md",
		mkPost("时间表", "aaa111", []string{"我的时间表很重要。", "时间表帮助我规划。", "另一个时间表出现。"}, true))
	// 论温柔：motions 用 /p/ 形态，作品集副本用 /album/ 形态，
	// 验证两种 Refer 形态下同一 postHash 仍能去重
	rouBody := []string{"温柔是一种力量。"}
	writeFixtureFile(t, dir, "alice/motions/2020-12-16_18_04_49_论温柔.md", mkPost("论温柔", "bbb222", rouBody, false))
	writeFixtureFile(t, dir, "alice/个人成长/2020-12-16_18_04_49_论温柔.md", mkAlbumPost("论温柔", "bbb222", rouBody, false))
	writeFixtureFile(t, dir, "alice/亲子教育/2020-12-16_18_04_49_论温柔.md", mkAlbumPost("论温柔", "bbb222", rouBody, false))
	writeFixtureFile(t, dir, "alice/motions/2021-06-01_09_00_00_信念与执念.md",
		mkPost("信念与执念", "ccc333", []string{"信念支撑我们。", "执念却是另一回事。", "信仰更深入。"}, false))
	writeFixtureFile(t, dir, "alice/motions/2021-07-01_09_00_00_信念之书.md",
		mkPost("信念之书", "eee555", []string{"信念之书记录信念。"}, false))

	// bob：仅作品集，无 motions（验证 canonical 回退）
	writeFixtureFile(t, dir, "bob/灯神/2019-05-05_00_00_00_只在专辑.md",
		mkPost("只在专辑", "ddd444", []string{"专辑内容。"}, false))

	return dir
}

// ---- 调用辅助 ----

func callTool(t *testing.T, h func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) *mcp.CallToolResult {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args
	res, err := h(context.Background(), req)
	if err != nil {
		t.Fatalf("handler 返回 err: %v", err)
	}
	return res
}

func findPost(posts []storage.Post, title string) (storage.Post, bool) {
	for _, p := range posts {
		if p.Title == title {
			return p, true
		}
	}
	return storage.Post{}, false
}

// ---- parseQuery ----

func TestParseQuery(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"信念 执念", []string{"信念", "执念"}},
		{`"时间 表"`, []string{"时间 表"}},
		{"foo, bar; baz", []string{"foo", "bar", "baz"}},
		{"   ", nil},
		{"信念 信念", []string{"信念"}},
		{`"a b" c`, []string{"a b", "c"}},
		{"FOO", []string{"foo"}},
	}
	for _, c := range cases {
		got := parseQuery(c.in)
		if len(got) == 0 && len(c.want) == 0 {
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseQuery(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// ---- 索引 ----

func TestBuildAuthorIndex_Dedup(t *testing.T) {
	dir := setupFixture(t)
	posts, err := storage.BuildAuthorIndex(dir, "alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 4 {
		t.Fatalf("alice 应有 4 篇去重文章，实际 %d", len(posts))
	}

	p, ok := findPost(posts, "论温柔")
	if !ok {
		t.Fatal("未找到 论温柔")
	}
	if p.ID != "bbb222" {
		t.Errorf("论温柔 id = %q, want bbb222", p.ID)
	}
	if want := []string{"个人成长", "亲子教育"}; !reflect.DeepEqual(p.Collections, want) {
		t.Errorf("论温柔 collections = %v, want %v", p.Collections, want)
	}
	if !strings.Contains(p.CanonicalPath, "/motions/") {
		t.Errorf("论温柔 canonical 应在 motions/，实际 %q", p.CanonicalPath)
	}
	if p.URL != "https://afdian.com/p/bbb222" {
		t.Errorf("论温柔 url = %q", p.URL)
	}

	// 日期倒序
	if posts[0].Title != "时间表" {
		t.Errorf("首篇应为最新的 时间表，实际 %q", posts[0].Title)
	}

	// 字数与评论标记
	shi, _ := findPost(posts, "时间表")
	if !shi.HasComments {
		t.Error("时间表 应标记有评论")
	}
	if shi.WordCount == 0 {
		t.Error("时间表 字数不应为 0")
	}
	if rou, _ := findPost(posts, "论温柔"); rou.HasComments {
		t.Error("论温柔 不应标记有评论")
	}
}

func TestBuildAuthorIndex_CanonicalFallback(t *testing.T) {
	dir := setupFixture(t)
	posts, err := storage.BuildAuthorIndex(dir, "bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("bob 应有 1 篇，实际 %d", len(posts))
	}
	p := posts[0]
	if p.ID != "ddd444" {
		t.Errorf("id = %q, want ddd444", p.ID)
	}
	if p.CanonicalPath != "bob/灯神/2019-05-05_00_00_00_只在专辑.md" {
		t.Errorf("无 motions 时 canonical 应回退到作品集路径，实际 %q", p.CanonicalPath)
	}
	if want := []string{"灯神"}; !reflect.DeepEqual(p.Collections, want) {
		t.Errorf("collections = %v, want %v", p.Collections, want)
	}
}

func TestBuildIndex_All(t *testing.T) {
	dir := setupFixture(t)
	posts, err := storage.BuildIndex(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 5 {
		t.Fatalf("全量应有 5 篇，实际 %d", len(posts))
	}
}

func TestExtractHash(t *testing.T) {
	cases := map[string]string{
		"https://afdian.com/p/xyz123":                "xyz123",
		"https://afdian.com/album/albumABC/postXYZ":  "postXYZ", // 作品集形态取末段
		"xyz123":                                     "xyz123",
		"  abc  ":                                    "abc",
	}
	for in, want := range cases {
		if got := storage.ExtractHash(in); got != want {
			t.Errorf("ExtractHash(%q) = %q, want %q", in, got, want)
		}
	}
}

// ---- 评论解析 ----

func TestParseComments(t *testing.T) {
	content := mkPost("t", "h", []string{"正文"}, true)
	comments := storage.ParseComments(content)
	if len(comments) != 3 {
		t.Fatalf("应解析 3 条评论，实际 %d", len(comments))
	}
	if !comments[0].IsHot || comments[0].Commenter != "张三" || comments[0].Text != "热评内容一" {
		t.Errorf("第一条热评解析错误：%+v", comments[0])
	}
	if comments[0].Time != "2021-01-01 10:00:00" {
		t.Errorf("时间解析错误：%q", comments[0].Time)
	}
	if comments[2].IsHot || comments[2].Commenter != "王五" {
		t.Errorf("普通评论解析错误：%+v", comments[2])
	}
}

// ---- search ----

func TestSearch_AND_Multiword(t *testing.T) {
	dir := setupFixture(t)

	// P0 回归：多词查询不再静默返回空
	out, err := Search(dir, searchParams{query: "信念 执念", matchAll: true, author: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if out.TotalDocs != 1 || out.Results[0].ID != "ccc333" {
		t.Fatalf("信念 执念(AND) 应命中 ccc333 一篇，实际 %d 篇: %+v", out.TotalDocs, out.Results)
	}

	// 三词 AND 仍命中（标题+正文覆盖）
	out, _ = Search(dir, searchParams{query: "信念 执念 信仰", matchAll: true, author: "alice"})
	if out.TotalDocs != 1 || out.Results[0].ID != "ccc333" {
		t.Fatalf("信念 执念 信仰(AND) 应命中 ccc333，实际 %+v", out.Results)
	}

	// AND 中有不存在的词 → 0
	out, _ = Search(dir, searchParams{query: "信念 不存在词XYZ", matchAll: true, author: "alice"})
	if out.TotalDocs != 0 {
		t.Fatalf("含不存在词的 AND 应为 0，实际 %d", out.TotalDocs)
	}
}

func TestSearch_Any(t *testing.T) {
	dir := setupFixture(t)
	out, _ := Search(dir, searchParams{query: "时间表 信念", matchAll: false, author: "alice"})
	if out.TotalDocs != 3 {
		t.Fatalf("时间表 信念(OR) 应命中 3 篇，实际 %d", out.TotalDocs)
	}
	out, _ = Search(dir, searchParams{query: "时间表 信念", matchAll: true, author: "alice"})
	if out.TotalDocs != 0 {
		t.Fatalf("时间表 信念(AND) 应为 0，实际 %d", out.TotalDocs)
	}
}

func TestSearch_Dedup(t *testing.T) {
	dir := setupFixture(t)
	out, _ := Search(dir, searchParams{query: "温柔", matchAll: true, author: "alice"})
	if out.TotalDocs != 1 {
		t.Fatalf("温柔 应去重为 1 篇（而非 motions+2 作品集 共 3 条），实际 %d", out.TotalDocs)
	}
	if out.Results[0].HitCount == 0 {
		t.Error("hit_count 不应为 0")
	}
}

func TestSearch_RankingAndPagination(t *testing.T) {
	dir := setupFixture(t)
	// 信念 命中 ccc333(2021-06) 与 eee555(2021-07)，覆盖度相同 → 按日期新近排序
	out, _ := Search(dir, searchParams{query: "信念", matchAll: true, author: "alice", limit: 1, offset: 0})
	if out.TotalDocs != 2 {
		t.Fatalf("信念 应命中 2 篇，实际 %d", out.TotalDocs)
	}
	if out.ReturnedDocs != 1 || out.Results[0].ID != "eee555" {
		t.Fatalf("limit=1 应返回最新的 eee555，实际 %+v", out.Results)
	}
	if out.NextOffset == nil || *out.NextOffset != 1 {
		t.Fatalf("next_offset 应为 1，实际 %v", out.NextOffset)
	}
	// 翻页
	out2, _ := Search(dir, searchParams{query: "信念", matchAll: true, author: "alice", limit: 1, offset: 1})
	if out2.Results[0].ID != "ccc333" {
		t.Fatalf("offset=1 应返回 ccc333，实际 %+v", out2.Results)
	}
	if out2.NextOffset != nil {
		t.Errorf("末页 next_offset 应为 nil，实际 %v", *out2.NextOffset)
	}
}

func TestSearch_PerDocHits(t *testing.T) {
	dir := setupFixture(t)
	out, _ := Search(dir, searchParams{query: "时间表", matchAll: true, author: "alice", perDocHits: 2})
	if out.TotalDocs != 1 {
		t.Fatalf("时间表 应命中 1 篇，实际 %d", out.TotalDocs)
	}
	r := out.Results[0]
	// 标题 heading 行 "## 时间表" + 3 个正文行 = 4 处命中
	if r.HitCount != 4 {
		t.Errorf("时间表 hit_count 应为 4，实际 %d", r.HitCount)
	}
	if len(r.Snippets) != 2 {
		t.Errorf("per_doc_hits=2 应返回 2 个片段，实际 %d", len(r.Snippets))
	}
}

func TestSearch_CommentHit(t *testing.T) {
	dir := setupFixture(t)
	out, _ := Search(dir, searchParams{query: "热评内容", matchAll: true, author: "alice"})
	if out.TotalDocs != 1 || out.Results[0].ID != "aaa111" {
		t.Fatalf("应能在评论区命中 aaa111，实际 %+v", out.Results)
	}
}

func TestSearch_CollectionFilter(t *testing.T) {
	dir := setupFixture(t)
	out, _ := Search(dir, searchParams{query: "温柔", matchAll: true, author: "alice", collection: "个人成长"})
	if out.TotalDocs != 1 {
		t.Fatalf("collection=个人成长 应命中 1 篇，实际 %d", out.TotalDocs)
	}
	out, _ = Search(dir, searchParams{query: "时间表", matchAll: true, author: "alice", collection: "个人成长"})
	if out.TotalDocs != 0 {
		t.Fatalf("时间表 不属于 个人成长，应为 0，实际 %d", out.TotalDocs)
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	dir := setupFixture(t)
	if _, err := Search(dir, searchParams{query: "   ", author: "alice"}); err == nil {
		t.Error("空白查询应返回错误")
	}
}

// ---- list_posts handler ----

func TestHandleListPosts_Pagination(t *testing.T) {
	dir := setupFixture(t)
	h := handleListPosts(dir)

	res := callTool(t, h, map[string]any{"author": "alice", "limit": 2, "offset": 0})
	out := res.StructuredContent.(postsOut)
	if out.TotalCount != 4 || out.Returned != 2 {
		t.Fatalf("total=4 returned=2，实际 total=%d returned=%d", out.TotalCount, out.Returned)
	}
	if out.Posts[0].Title != "时间表" || out.Posts[1].Title != "信念之书" {
		t.Errorf("首页顺序错误：%s, %s", out.Posts[0].Title, out.Posts[1].Title)
	}
	if out.NextOffset == nil || *out.NextOffset != 2 {
		t.Fatalf("next_offset 应为 2")
	}

	res = callTool(t, h, map[string]any{"author": "alice", "limit": 2, "offset": 2})
	out = res.StructuredContent.(postsOut)
	if out.NextOffset != nil {
		t.Error("末页 next_offset 应为 nil")
	}
}

func TestHandleListPosts_Filters(t *testing.T) {
	dir := setupFixture(t)
	h := handleListPosts(dir)

	out := callTool(t, h, map[string]any{"author": "alice", "collection": "个人成长"}).StructuredContent.(postsOut)
	if out.TotalCount != 1 || out.Posts[0].Title != "论温柔" {
		t.Fatalf("collection 过滤错误：%+v", out.Posts)
	}

	out = callTool(t, h, map[string]any{"author": "alice", "title_contains": "信念"}).StructuredContent.(postsOut)
	if out.TotalCount != 2 {
		t.Fatalf("title_contains=信念 应 2 篇，实际 %d", out.TotalCount)
	}

	out = callTool(t, h, map[string]any{"author": "alice", "date_from": "2021-01-01"}).StructuredContent.(postsOut)
	if out.TotalCount != 3 {
		t.Fatalf("date_from 过滤应 3 篇，实际 %d", out.TotalCount)
	}

	out = callTool(t, h, map[string]any{"author": "alice", "date_to": "2021-06-30"}).StructuredContent.(postsOut)
	if out.TotalCount != 2 {
		t.Fatalf("date_to 过滤应 2 篇，实际 %d", out.TotalCount)
	}
}

// ---- read_post handler ----

func TestHandleReadPost_ByIDURLPath(t *testing.T) {
	dir := setupFixture(t)
	h := handleReadPost(dir)

	for _, args := range []map[string]any{
		{"id": "ccc333"},
		{"url": "https://afdian.com/p/ccc333"},
		{"url": "ccc333"},
	} {
		out := callTool(t, h, args).StructuredContent.(readPostOut)
		if out.ID != "ccc333" || !strings.Contains(out.Content, "信念支撑我们") {
			t.Errorf("args=%v 定位错误：id=%q", args, out.ID)
		}
	}

	out := callTool(t, h, map[string]any{"path": "alice/motions/2022-03-02_10_00_00_时间表.md"}).StructuredContent.(readPostOut)
	if out.ID != "aaa111" || !strings.Contains(out.Content, "时间表帮助我规划") {
		t.Errorf("path 模式定位错误：%+v", out.ID)
	}
}

func TestHandleReadPost_ByTitle(t *testing.T) {
	dir := setupFixture(t)
	h := handleReadPost(dir)

	// 唯一标题 → 返回正文
	out := callTool(t, h, map[string]any{"author": "alice", "title": "时间表"}).StructuredContent.(readPostOut)
	if out.ID != "aaa111" {
		t.Errorf("唯一标题应直接返回，实际 id=%q", out.ID)
	}

	// 多义标题 → 消歧列表
	dis := callTool(t, h, map[string]any{"author": "alice", "title": "信念"}).StructuredContent.(disambiguationOut)
	if len(dis.Candidates) != 2 {
		t.Fatalf("信念 应返回 2 个候选，实际 %d", len(dis.Candidates))
	}
}

func TestHandleReadPost_MissingArgs(t *testing.T) {
	dir := setupFixture(t)
	res := callTool(t, handleReadPost(dir), map[string]any{})
	if !res.IsError {
		t.Error("缺少全部定位参数应返回错误")
	}
}

// ---- list_comments handler ----

func TestHandleListComments(t *testing.T) {
	dir := setupFixture(t)
	h := handleListComments(dir)

	out := callTool(t, h, map[string]any{"id": "aaa111"}).StructuredContent.(commentsOut)
	if out.TotalCount != 3 {
		t.Fatalf("aaa111 应有 3 条评论，实际 %d", out.TotalCount)
	}

	out = callTool(t, h, map[string]any{"id": "aaa111", "hot_only": true}).StructuredContent.(commentsOut)
	if out.TotalCount != 2 {
		t.Fatalf("hot_only 应 2 条，实际 %d", out.TotalCount)
	}

	out = callTool(t, h, map[string]any{"id": "aaa111", "commenter": "张三"}).StructuredContent.(commentsOut)
	if out.TotalCount != 1 {
		t.Fatalf("commenter=张三 应 1 条，实际 %d", out.TotalCount)
	}

	// path 模式
	out = callTool(t, h, map[string]any{"path": "alice/motions/2022-03-02_10_00_00_时间表.md"}).StructuredContent.(commentsOut)
	if out.TotalCount != 3 {
		t.Fatalf("path 模式应 3 条，实际 %d", out.TotalCount)
	}

	// 无评论文章
	out = callTool(t, h, map[string]any{"id": "bbb222"}).StructuredContent.(commentsOut)
	if out.TotalCount != 0 {
		t.Fatalf("bbb222 应 0 条，实际 %d", out.TotalCount)
	}

	if !callTool(t, h, map[string]any{}).IsError {
		t.Error("缺少定位参数应返回错误")
	}
}

// ---- list_authors handler ----

func TestHandleListAuthors(t *testing.T) {
	dir := setupFixture(t)
	out := callTool(t, handleListAuthors(dir), map[string]any{}).StructuredContent.(authorsOut)
	if out.Total != 2 {
		t.Fatalf("应有 2 位作者，实际 %d", out.Total)
	}
	counts := map[string]int{}
	for _, a := range out.Authors {
		counts[a.Author] = a.PostCount
	}
	if counts["alice"] != 4 || counts["bob"] != 1 {
		t.Errorf("文章数错误：%v", counts)
	}
}
