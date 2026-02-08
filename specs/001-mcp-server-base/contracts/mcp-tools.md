# MCP Tool Contracts

**Feature**: 001-mcp-server-base
**Date**: 2026-02-08

## Server Info

- **Name**: `AfdianToMarkdown`
- **Version**: same as CLI version
- **Transport**: stdio
- **Capabilities**: Tools (listChanged: false)

---

## Tool: `list_authors`

**Description**: 列出数据目录下所有已下载的作者

### Input Schema

```json
{
  "type": "object",
  "properties": {},
  "required": []
}
```

### Output (text)

```
已下载的作者（共 N 位）：
- author1
- author2
- ...
```

### Error Cases

| Condition            | Response                      |
|----------------------|-------------------------------|
| 数据目录为空          | `"当前没有已下载的作者。"`     |
| 数据目录不存在        | `"数据目录不存在：{path}"`    |

---

## Tool: `list_posts`

**Description**: 列出指定作者下的所有文章，按动态和作品集分组

### Input Schema

```json
{
  "type": "object",
  "properties": {
    "author": {
      "type": "string",
      "description": "作者的 URL slug（即目录名）"
    }
  },
  "required": ["author"]
}
```

### Output (text)

```
作者：{author}

## 动态（共 M 篇）
- [2020-12-15] 一个小打算 → q9adg/motions/2020-12-15_20_10_39_一个小打算.md
- [2020-12-16] 论温柔 → q9adg/motions/2020-12-16_18_04_49_论温柔.md
- ...

## 作品集：个人成长（共 K 篇）
- [2024-03-07] 那些被父母催婚逼婚的... → q9adg/个人成长/2024-03-07_...md
- ...

## 作品集：亲密关系（共 J 篇）
- ...
```

### Error Cases

| Condition          | Response                           |
|--------------------|------------------------------------|
| 作者不存在          | `"作者不存在：{author}"`            |
| 作者目录下无文章    | `"作者 {author} 下没有任何文章。"` |

---

## Tool: `read_post`

**Description**: 读取指定文章的完整 Markdown 内容

### Input Schema

```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "文章的相对路径（相对于数据目录）"
    },
    "author": {
      "type": "string",
      "description": "作者名（与 title 配合使用）"
    },
    "title": {
      "type": "string",
      "description": "文章标题关键词（模糊匹配）"
    }
  },
  "required": []
}
```

**参数使用规则**：
- 提供 `path` 时直接读取该文件
- 提供 `author` + `title` 时按标题关键词模糊匹配
- 两者都未提供时返回错误

### Output (text)

直接返回 Markdown 文件的完整内容，前面附加元信息行：

```
📄 文件：{relative_path}

{file content}
```

### Error Cases

| Condition              | Response                                     |
|------------------------|----------------------------------------------|
| path 指向不存在的文件   | `"文件不存在：{path}"`                        |
| 未提供 path 也未提供 author+title | `"请提供 path 或 author+title 参数"` |
| title 匹配到多篇文章    | 返回匹配列表，格式如 list_posts               |
| title 无匹配            | `"未找到标题包含 '{title}' 的文章"`           |

---

## Tool: `search`

**Description**: 在已下载文档中全文搜索关键词

### Input Schema

```json
{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "搜索关键词"
    },
    "author": {
      "type": "string",
      "description": "限定搜索范围的作者名（可选）"
    }
  },
  "required": ["query"]
}
```

### Output (text)

```
搜索 "关键词" 的结果（显示 N/M 条）：

---
📄 q9adg/motions/2022-04-16_疫情那一篇.md（第 15 行）

  13 | 前面的上下文行
  14 | 前面的上下文行
> 15 | 包含**关键词**的匹配行
  16 | 后面的上下文行
  17 | 后面的上下文行

---
📄 q9adg/个人成长/2024-03-07_标题.md（第 8 行）

   6 | ...
   7 | ...
>  8 | 另一个匹配行
   9 | ...
  10 | ...

还有 X 条结果未显示。
```

### Error Cases

| Condition           | Response                              |
|---------------------|---------------------------------------|
| query 为空           | `"请提供搜索关键词"`                   |
| 无匹配结果           | `"未找到包含 '{query}' 的内容。"`      |
| 指定的 author 不存在  | `"作者不存在：{author}"`              |

### Limits

- 最多返回 20 条匹配结果
- 每条结果包含匹配行及前后各 3 行上下文
- 搜索为大小写不敏感的纯文本匹配
