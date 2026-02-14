# Tasks: MCP HTTP Transport

**Input**: Design documents from `/specs/002-mcp-http-transport/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Not requested in feature specification. No test tasks included.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: No new project setup needed — this feature modifies 2 existing files only. Phase 1 is empty.

**Checkpoint**: Existing project structure is ready. Proceed directly to implementation.

---

## Phase 2: User Story 1 - 以HTTP模式启动MCP服务 (Priority: P1) 🎯 MVP

**Goal**: 用户可以通过 `mcp --http` 启动 HTTP Streamable MCP 服务器，监听指定地址，远程 MCP 客户端可通过 HTTP 调用所有工具。

**Independent Test**: 运行 `./AfdianToMarkdown --dir ./data mcp --http`，然后用 curl 发送 JSON-RPC initialize 请求到 `http://localhost:8080/mcp`，验证返回正确的 capabilities 响应。

### Implementation for User Story 1

- [x] T001 [P] [US1] 在 mcp/server.go 中添加 `ServeHTTP(s *server.MCPServer, addr string) error` 函数，使用 `server.NewStreamableHTTPServer(s)` 创建 HTTP 传输层，通过 goroutine 调用 `httpServer.Start(addr)` 启动服务，使用 `signal.NotifyContext` 监听 SIGTERM/SIGINT 信号，收到信号后调用 `httpServer.Shutdown(ctx)` 优雅关闭（10秒超时）
- [x] T002 [P] [US1] 在 main.go 的 mcp 子命令中添加 `--http` (BoolFlag) 和 `--addr` (StringFlag, 默认 "0.0.0.0:8080") 两个标志，更新 Usage 描述为 "以 MCP Server 模式启动，通过 stdio 或 HTTP 提供文档检索服务"
- [x] T003 [US1] 在 main.go 的 mcp 子命令 Action 中添加条件分支：当 `cmd.Bool("http")` 为 true 时调用 `mcpserver.ServeHTTP(s, cmd.String("addr"))`，否则保持现有 `mcpserver.Serve(s)` 调用
- [x] T004 [US1] 手动验证：构建项目 `go build -o AfdianToMarkdown .`，运行 `./AfdianToMarkdown --dir ./data mcp --http`，确认日志输出监听地址，用 curl 发送 initialize 请求验证 HTTP 端点响应正常

**Checkpoint**: HTTP MCP 服务可启动、可接受连接、可调用工具。MVP 完成。

---

## Phase 3: User Story 2 - 优雅关闭HTTP服务 (Priority: P2)

**Goal**: 服务器在收到 SIGTERM/SIGINT 信号后优雅关闭，等待当前请求处理完成。

**Independent Test**: 启动 HTTP 服务，发送 SIGTERM 信号，验证服务器日志输出关闭消息并正常退出（退出码 0）。

### Implementation for User Story 2

> 注意：优雅关闭的核心逻辑已在 T001 中通过 `signal.NotifyContext` + `Shutdown(ctx)` 实现。本阶段仅需验证行为正确。

- [x] T005 [US2] 手动验证优雅关闭：启动 HTTP 服务后发送 `kill -SIGTERM <pid>`，确认日志输出 "正在关闭 MCP HTTP Server..."，进程正常退出；发送 `kill -SIGINT <pid>`（或 Ctrl+C），确认相同行为

**Checkpoint**: 优雅关闭在两种信号下均正常工作。

---

## Phase 4: User Story 3 - 使用现有stdio模式（向后兼容） (Priority: P3)

**Goal**: 不带 `--http` 标志时，mcp 子命令行为与之前完全一致。

**Independent Test**: 运行 `./AfdianToMarkdown --dir ./data mcp`（不加 --http），验证以 stdio 模式启动。

### Implementation for User Story 3

> 注意：向后兼容已在 T003 的条件分支中保证。本阶段仅需验证。

- [x] T006 [US3] 手动验证向后兼容：运行 `./AfdianToMarkdown --dir ./data mcp`（不带 --http），确认以 stdio 模式启动，日志输出 "MCP Server 已就绪，等待连接..."

**Checkpoint**: stdio 模式行为不变，向后兼容确认。

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: 构建验证和文档

- [x] T007 确认 `go build` 无编译错误，`go vet ./...` 无警告
- [x] T008 运行 quickstart.md 中的 curl 验证步骤，确认 initialize 和 tools/call 请求均返回正确响应

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 2 (US1)**: No dependencies — can start immediately
- **Phase 3 (US2)**: Depends on T001 (graceful shutdown is part of ServeHTTP implementation)
- **Phase 4 (US3)**: Depends on T003 (conditional branch preserves stdio path)
- **Phase 5 (Polish)**: Depends on all user stories complete

### Within User Story 1

- T001 and T002 can run in **parallel** (different files: mcp/server.go vs main.go)
- T003 depends on both T001 and T002 (wires them together in main.go Action)
- T004 depends on T003 (end-to-end verification)

### Parallel Opportunities

```
T001 (mcp/server.go) ──┐
                        ├── T003 (main.go Action) ── T004 (验证)
T002 (main.go Flags) ──┘
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Implement T001 + T002 in parallel (two different files)
2. Wire together in T003
3. Build and verify with T004
4. **STOP and VALIDATE**: curl test confirms HTTP MCP works

### Incremental Delivery

1. US1 (T001-T004) → HTTP 服务可用 → MVP!
2. US2 (T005) → 验证优雅关闭
3. US3 (T006) → 验证向后兼容
4. Polish (T007-T008) → 代码质量和文档验证

---

## Notes

- 本特性仅修改 2 个文件：`mcp/server.go` 和 `main.go`
- 工具处理器（tools.go）、搜索逻辑（search.go）、存储层（storage/）均无需修改
- T001 中的 `ServeHTTP` 函数已包含优雅关闭逻辑，US2 本质上是验证而非新增实现
- `--addr` 参数使用 `host:port` 格式避免与全局 `--host`（afdian 域名）冲突
