# MCP 工具重构设计：面向 agent 的检索能力

日期：2026-06-27
状态：已确认（待实现）
范围：`mcp/`、`storage/`

## 背景与问题

当前 MCP Server 暴露 4 个 tool（`list_authors`、`list_posts`、`read_post`、`search`），实测对 agent 消费不友好，已知问题：

- **P0 — search 不分词（隐形杀手）**：`信念 执念 信仰` → 0 条，但 `信念` → 218 条。根因是把含空格的整串当字面子串匹配（`search.go` 的 `strings.Contains(line, queryLower)`）。后果：任何多词查询静默返回空，agent 误判为"没有"而漏掉。
- **P1 — 取不全 + 无排序 + 不跨文档去重**：搜 `时间表` 返回"显示 20/177"，但①无 offset/limit 取后 157 条；②20 条几乎全来自同一篇、按行号顺序倾泻，把别的文档挤没。
- **P2 — 缺结构化输出 + 无法按 id/url 取文**：
  - `list_posts` 一次吐 1867 行 / 18 万字直接爆 token。
  - `read_post` 按标题命中"3 篇请指定路径"，根因是同一篇被归进 motions + 多个作品集多份。
- **加分项**：评论区是金矿（热评里有最戳的读者自白，且 search 已能命中评论），希望有 `list_comments` / 按评论者过滤；列表直接带出 Refer 原始 URL 方便回链。

## 根因诊断

所有问题共享一个根因：**没有文档模型**。当前所有 tool 直接 `filepath.Walk` 裸文件 + 字面子串匹配，导致：

- 同一篇文章在 `motions/` 与每个所属作品集目录各存一份（文件名相同、`### Refer` 的 hash 相同），无去重概念 → read_post 多命中、search 双命中、list_posts 重复列出。
- 没有"词项"概念 → 查询串整体匹配。
- 没有"结果集"概念 → 无分页、无排序、无投影。

**经验证的数据事实**（作者 `alice`）：

- 每个 `.md` 文件头部都有 `### Refer` 段，含 `https://afdian.com/p/<hash>`，**100% 覆盖**（`writer.go` 无条件写入），`<hash>` 即 canonical id。
- `motions/` 是完整超集：839 个唯一 Refer hash = 该作者全部唯一 hash。同一篇在 album 与 motions 下**文件名一致、hash 一致**。
- 注意：仅用 `albums` 命令下载的作者**没有 `motions/`**，canonical 路径选择需回退。
- 文件名格式 `YYYY-MM-DD_HH_MM_SS_<安全标题>.md`，日期可从文件名解析。
- 评论区：`## 热评`（热评）/`## 评论`（普通评论）两段；每条 `##### <span>[N] YYYY-MM-DD HH:MM:SS by 评论者</span>`，`----` 分隔，正文在 header 后。

## 已确认的设计决策

1. **范围**：一份完整 spec，按 P0 → P1 → P2 → 加分项 优先级实现。
2. **架构**：引入共享的**文档索引层**（`storage`），把 `data/` 解析为按 canonical id 去重的 `Post` 列表，所有 tool 查询它。
3. **索引生命周期**：**每次调用按需构建，无缓存**（始终新鲜，无失效问题）。痛点是正确性而非性能；几百到上千文件全读一遍为毫秒级。若将来长驻服务出现性能问题再加缓存。
4. **输出格式**：每个 tool **紧凑 JSON 为主 + 人读文本镜像**。落地用 `mcp.NewToolResultStructured(jsonObj, humanText)`（structuredContent = JSON 对象，fallbackText = 人读渲染），并配 `WithOutputSchema[T]()`。
5. **搜索排序**：**简单启发式**——AND 覆盖度（命中的不同词数）→ 总命中次数 → 标题命中加权 → 日期新近。不上 TF-IDF/BM25。
6. **搜索匹配语义**：whitespace/标点切分；`"..."` 段为短语精确子串；裸词默认 **AND**；`match` 参数（`all`|`any`，默认 `all`）支持 OR；大小写不敏感；**子串匹配，不做中文分词**。tool 描述显式引导"多概念请用空格分隔"。
7. **不引入分词库**：whitespace 切分已覆盖 P0（空格串）；无空格中文串靠 tool 描述引导规避；消费方是 agent，可控。升级路径若需要原生支持无空格中文串，用纯 Go 的 `go-ego/gse`（**不要** cgo 的 gojieba，会破坏 goreleaser 跨平台编译）。
8. **read_post**：把 by-url 折叠进 `read_post` 的 `url` 参数（不单开 `get_post_by_url`）。
9. **list_authors**：顺带返回每位作者的文章数（去重后的 canonical id 数）。
10. **word_count**：正文 rune（字符）计数，剔除图片/媒体标签，为估算值。

## 架构

### 数据流（重构后）

```
MCP tool 调用
  → storage.BuildIndex(dataDir[, author])   // 每次调用构建，按 Refer hash 去重
  → []Post（含 collections / canonical_path / url / word_count / has_comments）
  → tool handler：过滤 / 排序 / 分页 / 投影
  → mcp.NewToolResultStructured(JSON, 文本镜像)
```

### 文件改动

| 文件                    | 动作                                                                                                                                                                          |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `storage/index.go`    | **新增**：`Post` 模型 + `BuildIndex` / `BuildAuthorIndex` + 查询/过滤/分页辅助                                                                                    |
| `storage/comments.go` | **新增**：评论解析 `ParseComments`                                                                                                                                    |
| `mcp/render.go`       | **新增**：各 tool 的 JSON 结构体 + 文本镜像渲染收口                                                                                                                     |
| `mcp/search.go`       | **改写**：查询分词、启发式排序、文档级结果、按 id 去重                                                                                                                  |
| `mcp/tools.go`        | **改写**：各 handler 接新参数、调索引、走结构化输出                                                                                                                     |
| `mcp/server.go`       | **改写**：注册新参数 + `list_comments` tool + 各 tool 的 OutputSchema                                                                                                 |
| `storage/reader.go`   | **重构**：`ListPosts`/`FindPostByTitle` 改为基于索引，消除重复逻辑；保留 `safePath`/`ReadPost`/`ParsePostInfo`/`ListAuthors`（仅 `mcp` 包消费，改造安全） |

## 组件设计

### 1. 文档索引层（`storage/index.go`）

```go
type Post struct {
    ID            string   // canonical id = ### Refer 里 afdian.com/p/<hash> 的 hash
    Title         string   // 文件名解析（复用 ParsePostInfo）
    Author        string
    Date          string   // YYYY-MM-DD（文件名解析）
    Collections   []string // 该 id 所属作品集名（不含 motions），排序
    CanonicalPath string   // 去重代表路径：优先 motions/，否则字母序首个 album
    URL           string   // https://afdian.com/p/<hash>
    WordCount     int      // 正文 rune 数（剔除图片/媒体标签），估算
    HasComments   bool     // 是否含 ## 热评 或 ## 评论
}
```

**构建算法**：

1. 遍历作者目录下所有 `.md`（跳过 `.assets`）。
2. 每个文件读头部数行，正则 `afdian\.com/p/([a-zA-Z0-9]+)` 取 Refer hash。
3. 按 `(author, hash)` 分组去重。记录每个路径所在子目录（category）。
4. 每组：
   - `CanonicalPath` = category 为 `motions` 的路径；无则字母序首个 album 路径。
   - `Collections` = 组内 category ≠ `motions` 的子目录名，去重排序。
   - 整读 `CanonicalPath` 一次，算 `WordCount`、`HasComments`、`URL`、`Title`、`Date`。
5. 返回按 Date 倒序的 `[]Post`。

接口：

```go
func BuildAuthorIndex(dataDir, author string) ([]Post, error) // 单作者
func BuildIndex(dataDir string) ([]Post, error)               // 全部作者（search 无 author 时 / list_authors 计数）
```

`WordCount` 算法：截取 `### 正文` 到下一个 `##`/`### ` 之间的文本，剔除 `![...](...)` 图片行与 `<audio>/<video>` 标签，统计非空白 rune 数。

### 2. 评论解析（`storage/comments.go`）

```go
type Comment struct {
    Index     int    // 段内序号 [N]
    Time      string // YYYY-MM-DD HH:MM:SS
    Commenter string // 评论者名
    Text      string // 评论正文
    IsHot     bool   // true=热评(## 热评)，false=普通(## 评论)
}

func ParseComments(content string) []Comment
```

解析：定位 `## 热评` / `## 评论` 段；段内以 `----` 分隔条目；header 正则
`^#{5}\s*<span>\[(\d+)\]\s+(\d{4}-\d\d-\d\d \d\d:\d\d:\d\d)\s+by\s+(.+?)</span>`；
正文取 header 后到下一个 `----`/`##`/EOF 的非空文本（trim）。

## Tool 契约

所有 tool：JSON 为主（structuredContent）+ 人读文本镜像（fallbackText）。空结果返回结构化空集（`total=0`），不报错。错误（参数缺失、作者/文章不存在、路径越界）走 `NewToolResultError`。保留 `safePath` 路径越界防护。

### search（P0 + P1）

**参数**：

| 参数             | 类型   | 默认    | 说明                                                |
| ---------------- | ------ | ------- | --------------------------------------------------- |
| `query`        | string | 必填    | 查询串；`"..."` 为短语精确子串；裸词空白/标点切分 |
| `match`        | string | `all` | `all`=AND，`any`=OR                             |
| `author`       | string | —      | 限定作者                                            |
| `collection`   | string | —      | 限定作品集                                          |
| `limit`        | number | 10      | 文档级返回上限                                      |
| `offset`       | number | 0       | 文档级偏移                                          |
| `per_doc_hits` | number | 3       | 每文档片段上限                                      |

**逻辑**：

1. 解析 query：抽取 `"..."` 短语，其余按空白/标点切分为词项；全部小写。
2. 构建索引（按 author/collection 过滤范围），**仅搜 canonical 文件**（天然按 id 去重）。
3. 每篇读 canonical 全文（含评论，保留命中评论能力），逐行匹配词项/短语。
4. 文档级聚合：`hit_count` = 命中行数；`coverage` = 命中的不同词项数。
   - `match=all`：要求 coverage = 词项总数才入选；`any`：≥1。
5. 排序（启发式）：`coverage` 降 → 标题是否命中词项 降 → `hit_count` 降 → `date` 降。`score` 字段为信息性数值。
6. 分页：按 limit/offset 截取文档；每文档取前 `per_doc_hits` 条命中行（含 ±2 行上下文）。

**输出 JSON**：

```json
{
  "query": "信念 执念",
  "match": "all",
  "total_docs": 42,
  "total_hits": 177,
  "returned_docs": 10,
  "offset": 0,
  "next_offset": 10,
  "results": [
    {
      "id": "8bf2...", "title": "...", "author": "alice", "date": "2021-07-19",
      "path": "alice/motions/2021-...md", "url": "https://afdian.com/p/8bf2...",
      "collections": ["个人成长"], "score": 3210, "hit_count": 12,
      "snippets": [{"line": 23, "text": "  21 | ...\n> 23 | ...命中...\n  25 | ..."}]
    }
  ]
}
```

`next_offset` 在还有更多时为下一偏移，否则 `null`。文本镜像：标题行 `搜索 "X"：N 篇 / M 次命中（显示 a-b）` + 每篇 `[date] title (hits) → path` + 顶部片段。

### list_posts（P2）

**参数**：

| 参数               | 类型   | 默认 | 说明                         |
| ------------------ | ------ | ---- | ---------------------------- |
| `author`         | string | 必填 | 作者 url slug                |
| `collection`     | string | —   | 限定作品集                   |
| `title_contains` | string | —   | 标题子串过滤（大小写不敏感） |
| `date_from`      | string | —   | `YYYY-MM-DD`，含           |
| `date_to`        | string | —   | `YYYY-MM-DD`，含           |
| `limit`          | number | 50   | 返回上限                     |
| `offset`         | number | 0    | 偏移                         |

**逻辑**：构建作者索引（已按 id 去重）→ 应用过滤 → 按 date 倒序 → 分页 → 投影。

**输出 JSON**：

```json
{
  "author": "alice", "total_count": 800, "returned": 50, "offset": 0, "next_offset": 50,
  "posts": [
    {"id":"...","title":"...","date":"2021-07-19","collections":["个人成长"],
     "canonical_path":"alice/motions/...md","url":"https://afdian.com/p/...",
     "word_count": 1234, "has_comments": true}
  ]
}
```

`total_count` 为过滤后、分页前的数量。文本镜像：`作者 X：共 N 篇（显示 a-b）` + 每篇精简行。

### read_post（P2）

**参数**（优先级从高到低，取第一个有值者）：

| 参数                   | 说明                                                                                                |
| ---------------------- | --------------------------------------------------------------------------------------------------- |
| `id`                 | canonical hash，精确定位（最稳）                                                                    |
| `url`                | 全 URL 或 hash，抽取 hash 后按 id 匹配（折叠 get_post_by_url）                                      |
| `path`               | 相对路径直接读（现有，保留 safePath 校验）                                                          |
| `author` + `title` | 作者内标题子串匹配，**按 id 去重**；1 篇直接返回；多个**不同 id**则返回带 id 的消歧列表 |

**输出 JSON**：

```json
{"id":"...","title":"...","author":"...","date":"...","collections":["..."],
 "canonical_path":"...","url":"...","word_count":1234,"has_comments":true,
 "content":"<完整 markdown>"}
```

文本镜像：`📄 <path>` + 完整 content。消歧列表场景输出候选 `{id,title,date,path}` 列表。

### list_comments（加分项，新 tool）

**参数**：

| 参数                        | 类型   | 默认       | 说明                          |
| --------------------------- | ------ | ---------- | ----------------------------- |
| `id` / `url` / `path` | string | 三选一必填 | 定位文章（同 read_post 解析） |
| `commenter`               | string | —         | 评论者名子串过滤              |
| `hot_only`                | bool   | false      | 仅热评                        |
| `limit`                   | number | 50         | 返回上限                      |
| `offset`                  | number | 0          | 偏移                          |

**逻辑**：定位文章 → 读全文 → `ParseComments` → 过滤（commenter / hot_only）→ 分页。

**输出 JSON**：

```json
{"id":"...","title":"...","path":"...","total_count":120,"returned":50,"offset":0,"next_offset":50,
 "comments":[{"index":0,"time":"2020-12-15 20:30:16","commenter":"ciciff","text":"...","is_hot":true}]}
```

### list_authors（增强）

**输出 JSON**：

```json
{"total":3,"authors":[{"author":"alice","post_count":839}]}
```

`post_count` = 去重后 canonical id 数。文本镜像：`已下载作者（共 N 位）` + 每行 `- author（M 篇）`。

## 错误处理

- 参数缺失 / 互斥未满足（如 read_post 四组都为空）→ `NewToolResultError`，文案明确。
- 作者 / 文章 / 路径不存在 → `NewToolResultError`。
- 路径越界 → 现有 `safePath` 拦截。
- 查询解析后无有效词项（如仅空白）→ `NewToolResultError`。
- 命中为空 → 结构化空集（非错误）。

## 测试策略

沿用 `t.TempDir()` 合成夹具树（构造含 motions + 多作品集重复、含 `### Refer`、含 `## 热评`/`## 评论` 的样例文件），表驱动覆盖：

- **查询分词**：空格串拆词、`"短语"` 精确、`match=all/any`、标点切分、纯空白报错。
- **索引去重**：motions + album 同一篇 → 1 个 Post；`Collections` 正确；`CanonicalPath` 选 motions；**无 motions 时回退**字母序首个 album。
- **排序**：coverage > hit_count > 标题命中 > 新近 的顺序断言。
- **search 分页**：limit/offset/`per_doc_hits`/`next_offset`；评论命中可被搜到。
- **list_posts 过滤分页**：collection / title_contains / date_from / date_to / limit / offset / total_count。
- **read_post**：by id、by url（全 URL 与裸 hash）、by path、by author+title 去重为 1、多 id 消歧列表。
- **评论解析**：热评/普通分段、字段解析、commenter 过滤、hot_only。
- **list_authors**：post_count 去重计数。

不依赖网络（区别于 `afdian_test.go` 的真实 API 测试）。

## 非目标（YAGNI）

- 不做中文分词 / 倒排索引 / BM25。
- 不做索引缓存（每次调用构建）。
- 不改下载侧（`writer.go` / afdian 客户端）。
- 不做图文混排等既有 TODO。

## 落地顺序（实现时）

1. `storage/index.go` + `reader.go` 重构（去重模型，所有 tool 的地基）。
2. P0/P1：`mcp/search.go` 改写 + `mcp/render.go`。
3. P2：`list_posts` / `read_post` 改写。
4. 加分项：`storage/comments.go` + `list_comments` + `list_authors` 增强。
5. `mcp/server.go` 注册参数与 OutputSchema；补齐测试。
