# Implementation Plan: MCP Server 基础框架

**Branch**: `001-mcp-server-base` | **Date**: 2026-02-08 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-mcp-server-base/spec.md`

## Summary

在现有 AfdianToMarkdown CLI 中添加 `mcp` 子命令，以 stdio 传输模式启动 MCP Server，注册 4 个文档检索 Tool（`list_authors`、`list_posts`、`read_post`、`search`），使 Claude Code 能够连接并检索已下载的爱发电 Markdown 文档。使用 `mark3labs/mcp-go` 作为 MCP SDK，复用现有的 `utils` 包文件扫描函数和 `storage/reader.go`（待实现）进行文件读取与搜索。

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: `urfave/cli/v3`（CLI 框架）、`mark3labs/mcp-go v0.43.2`（MCP SDK）
**Storage**: 文件系统（Markdown 文件，只读）
**Testing**: `go test`、`testing` 标准库
**Target Platform**: Windows / Linux / macOS（跨平台 CLI）
**Project Type**: Single project（在现有 CLI 中添加子命令）
**Performance Goals**: 搜索 800+ 篇文档 < 3 秒，列表操作即时响应
**Constraints**: 纯 Go（无 CGO），不引入外部服务依赖
**Scale/Scope**: ~800 篇 Markdown 文档，单用户本地使用

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution 尚未配置（仅为模板），无具体 Gate 需要检查。默认通过。

## Project Structure

### Documentation (this feature)

```text
specs/001-mcp-server-base/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (MCP Tool schemas)
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
AfdianToMarkdown/
├── main.go              # 添加 mcp 子命令，修改 Before 钩子
├── mcp/
│   ├── server.go        # MCP Server 创建、Tool 注册、启动逻辑
│   ├── tools.go         # 4 个 Tool handler 实现
│   └── search.go        # 全文关键词搜索逻辑
├── storage/
│   └── reader.go        # 文件读取：作者列表、文章列表、文章内容读取
├── config/
│   └── config.go        # 现有，无需修改
└── utils/
    └── utils.go         # 现有，复用 CheckAndListAuthors 等
```

**Structure Decision**: 新建 `mcp/` 包承载 MCP Server 逻辑，`storage/reader.go` 实现文件读取层供 MCP Tool handler 调用。遵循现有代码按包组织的惯例，不引入新的顶层目录结构。

## Complexity Tracking

无违规需要说明。
