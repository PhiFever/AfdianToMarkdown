# Research: MCP Server 基础框架

**Feature**: 001-mcp-server-base
**Date**: 2026-02-08

## R1: MCP Go SDK 选型

**Decision**: 使用 `github.com/mark3labs/mcp-go v0.43.2`

**Rationale**:
- 社区最广泛采用的 Go MCP SDK（400+ 依赖项目）
- 纯 Go 实现，无 CGO 依赖，适合跨平台构建
- API 成熟：`server.NewMCPServer()` + `s.AddTool()` + `server.ServeStdio()`
- 支持 stdio 传输，与 Claude Code 直接兼容

**Alternatives considered**:
- `github.com/modelcontextprotocol/go-sdk`：官方 SDK（Google 维护），较新但 API 风格不同，社区采用度较低
- 手写 JSON-RPC over stdio：工作量大，无必要

## R2: MCP 子命令与 Before 钩子的关系

**Decision**: 在 `Before` 钩子中条件判断，`mcp` 子命令跳过 Cookie 加载

**Rationale**:
- 当前 `Before` 钩子统一加载 Cookie，MCP 模式不需要 Cookie
- `urfave/cli/v3` 的 `Before` 钩子在所有子命令前执行，无法按子命令选择性跳过
- 最简方案：在 `Before` 中检查当前子命令名，若为 `mcp` 则只初始化 Config（dataDir），跳过 Cookie 加载

**Alternatives considered**:
- 将 Cookie 加载移到各子命令的 `Before` 中：侵入性更大，需修改 4 个子命令
- 使用子命令自己的 `Before`：`urfave/cli/v3` 支持子命令级 `Before`，但会与全局 `Before` 都执行

## R3: storage/reader.go 职责设计

**Decision**: `storage/reader.go` 提供 3 类函数：列表扫描、文件读取、全文搜索

**Rationale**:
- MCP Tool handler 需要薄薄一层业务逻辑，核心 I/O 操作放在 storage 层
- 复用现有 `utils.CheckAndListAuthors()` 的模式，但返回更丰富的结构化信息
- 搜索逻辑放在独立的 `mcp/search.go` 中，因为搜索属于 MCP 特有功能，不是通用存储操作

**函数设计**:
- `ListAuthors(dataDir string) ([]string, error)` — 封装 `utils.CheckAndListAuthors`
- `ListPosts(dataDir, author string) (*AuthorPosts, error)` — 扫描 motions + albums 目录
- `ReadPost(filePath string) (string, error)` — 读取单个 Markdown 文件
- `FindPostByTitle(dataDir, author, titleKeyword string) ([]PostInfo, error)` — 按标题关键词模糊匹配

## R4: 全文搜索实现方案

**Decision**: 逐文件扫描 + `strings.Contains`（大小写不敏感），第一阶段不建索引

**Rationale**:
- 800 篇 Markdown 文件，平均每篇 ~5KB，总量约 4MB
- 全量读取+扫描在内存中仅需几十毫秒，远低于 3 秒目标
- 无需引入 SQLite/向量搜索/倒排索引等额外复杂度
- `strings.Contains` 对中文关键词天然支持（UTF-8 字节匹配）

**Alternatives considered**:
- 启动时构建内存倒排索引：当前规模不需要
- SQLite FTS5：引入 CGO 依赖，破坏纯 Go 构建
- sqlite-vec 向量搜索：过度工程，详见前序讨论

## R5: Tool 返回内容格式

**Decision**: 统一使用纯文本返回（`mcp.NewToolResultText`），内容为人类可读格式

**Rationale**:
- MCP Tool 返回值最终由 Claude 消费，纯文本最简单且 Claude 理解无障碍
- `list_authors` 返回换行分隔的作者名
- `list_posts` 返回按类别分组的标题列表（含相对路径）
- `read_post` 返回完整 Markdown 内容
- `search` 返回匹配片段（含文件路径、行号、上下文）

**Alternatives considered**:
- JSON 格式返回：增加复杂度，Claude 处理纯文本同样高效
- 多 content block 返回：MCP 支持但无必要
