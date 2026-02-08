# Data Model: MCP Server 基础框架

**Feature**: 001-mcp-server-base
**Date**: 2026-02-08

## Entities

### PostInfo

表示一篇文章的元信息（不含内容），用于列表和搜索结果返回。

| Field       | Type     | Description                              |
|-------------|----------|------------------------------------------|
| Title       | string   | 文章标题（从文件名中提取，去除时间戳前缀） |
| Path        | string   | 相对于数据目录的文件路径                   |
| Category    | string   | 类别："motions" 或作品集名称               |
| PublishTime | string   | 发布时间（从文件名提取，格式 YYYY-MM-DD）  |

### AuthorPosts

表示一位作者下所有文章的分组结构。

| Field    | Type                    | Description                |
|----------|-------------------------|----------------------------|
| Author   | string                  | 作者名（目录名）           |
| Motions  | []PostInfo              | 动态列表                   |
| Albums   | map[string][]PostInfo   | 作品集名 → 文章列表        |

### SearchResult

表示一条搜索匹配结果。

| Field      | Type     | Description                              |
|------------|----------|------------------------------------------|
| FilePath   | string   | 相对于数据目录的文件路径                   |
| Title      | string   | 文章标题                                  |
| Author     | string   | 作者名                                    |
| LineNumber | int      | 匹配行号                                  |
| Context    | string   | 匹配行及前后各 3 行的上下文文本            |

### SearchResponse

表示搜索的完整返回。

| Field       | Type           | Description                         |
|-------------|----------------|-------------------------------------|
| Query       | string         | 搜索关键词                           |
| TotalCount  | int            | 总匹配数                             |
| Results     | []SearchResult | 返回的匹配结果（上限 20 条）          |
| Truncated   | bool           | 是否因超过上限而截断                  |

## Relationships

```
DataDir (root)
  └── Author (directory)
        ├── motions/ (directory)
        │     └── PostInfo (*.md files)
        └── {AlbumName}/ (directory, 1..N)
              └── PostInfo (*.md files)
```

## File Name Parsing

文件名格式：`{YYYY-MM-DD_HH_MM_SS}_{SafeTitle}.md`

解析规则：
- 前 19 个字符为时间戳（`2020-12-15_20_10_39`）
- 第 20 个字符为分隔符 `_`
- 第 21 个字符起为安全标题
- 日期部分取前 10 个字符（`2020-12-15`）作为 PublishTime

## Notes

- 所有路径使用 `/` 作为分隔符（跨平台兼容）
- `.assets/` 目录在列表中被过滤，不作为 Album 或文章
- 数据模型为只读，MCP Server 不修改文件系统
