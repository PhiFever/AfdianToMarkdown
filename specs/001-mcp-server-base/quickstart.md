# Quickstart: MCP Server 基础框架

**Feature**: 001-mcp-server-base

## Prerequisites

- Go 1.24+
- 已下载的爱发电文档（位于 `data/` 目录下）

## Build

```bash
go build -o AfdianToMarkdown.exe .
```

## Run MCP Server

```bash
# 使用默认数据目录（程序所在目录/data）
./AfdianToMarkdown mcp

# 指定数据目录
./AfdianToMarkdown mcp --dir /path/to/docs
```

## Configure Claude Code

在 Claude Code 的 MCP 配置文件中添加：

```json
{
  "mcpServers": {
    "afdian-docs": {
      "command": "/path/to/AfdianToMarkdown",
      "args": ["mcp", "--dir", "/path/to/downloaded/docs"]
    }
  }
}
```

## Available Tools

| Tool             | Description                    | Parameters                                              |
|------------------|--------------------------------|---------------------------------------------------------|
| `list_authors`   | 列出所有已下载的作者            | 无                                                       |
| `list_posts`     | 列出作者下的文章               | `author` (string, required)                              |
| `read_post`      | 读取文章内容                    | `path` (string, optional), `author` + `title` (optional) |
| `search`         | 全文关键词搜索                  | `query` (string, required), `author` (string, optional)  |

## Usage Examples (in Claude Code)

```
> 帮我看看已经下载了哪些作者的内容
  → Claude calls list_authors

> 列出 q9adg 的所有文章
  → Claude calls list_posts(author="q9adg")

> 帮我找一下关于"人生意义"的文章
  → Claude calls search(query="人生意义")

> 读一下这篇文章：q9adg/motions/2022-04-16_18_16_43_疫情那一篇.md
  → Claude calls read_post(path="q9adg/motions/2022-04-16_18_16_43_疫情那一篇.md")
```

## Verify Connection

启动 Claude Code 后，在对话中输入：

```
帮我列出已下载的作者
```

如果 MCP Server 连接成功，Claude 会调用 `list_authors` 并返回作者列表。
