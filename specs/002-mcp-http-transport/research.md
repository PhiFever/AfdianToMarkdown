# Research: MCP HTTP Transport

## R1: mcp-go HTTP Streamable Transport API

**Decision**: 使用 `server.NewStreamableHTTPServer()` + `Start()` / `Shutdown()` API

**Rationale**: mcp-go v0.43.2 原生提供 `StreamableHTTPServer` 类型，完整支持 MCP Specification 2025-03-26 的 Streamable HTTP 传输协议。API 简洁，与现有 `MCPServer` 实例无缝集成。

**Alternatives considered**:
- 手动实现 HTTP 处理：不必要，mcp-go 已提供完整实现
- 使用 SSE transport：已被 MCP 规范标记为 deprecated

**API 详情**:

```go
// 创建
httpServer := server.NewStreamableHTTPServer(mcpServer, opts...)

// 启动（阻塞）
httpServer.Start(":8080")

// 优雅关闭
httpServer.Shutdown(ctx)

// 也实现了 http.Handler 接口，可嵌入自定义 http.Server
```

**可用 Options**:
- `WithEndpointPath(path)` — 端点路径，默认 `/mcp`
- `WithStateLess(bool)` — 无状态模式（默认已是 stateless）
- `WithDisableStreaming(bool)` — 禁用 SSE 流式响应
- `WithStreamableHTTPServer(*http.Server)` — 注入自定义 http.Server
- `WithHTTPContextFunc(fn)` — 自定义请求上下文

## R2: 优雅关闭方案

**Decision**: 使用 `signal.NotifyContext()` + `StreamableHTTPServer.Shutdown(ctx)`

**Rationale**: Go 标准库的 signal.NotifyContext 提供了最简洁的信号处理方式，与 context 链自然集成。`StreamableHTTPServer.Shutdown()` 内部调用 `http.Server.Shutdown()`，会等待活跃连接完成。

**实现模式**:
```go
ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stop()

// 在 goroutine 中启动 HTTP server
go httpServer.Start(addr)

<-ctx.Done()  // 等待信号

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
httpServer.Shutdown(shutdownCtx)
```

## R3: CLI 标志设计

**Decision**: 在 `mcp` 子命令下添加 `--http`、`--port`、`--host` 三个标志

**Rationale**:
- `--http` 布尔标志明确区分 stdio/HTTP 模式，保持向后兼容
- `--port` 和 `--host` 仅在 HTTP 模式下有意义
- 注意：`--host` 与全局的 `--host`（afdian 域名）冲突，需要改用 `--addr` 或 `--listen` 来避免歧义

**冲突解决**: 全局 `--host` 用于 afdian 域名，MCP 子命令的监听地址改用 `--addr` 参数，格式为 `host:port`（如 `--addr 0.0.0.0:8080`），或者分别使用 `--listen` 和 `--port`。

**最终方案**: 使用 `--addr` 单一参数，默认值 `0.0.0.0:8080`，格式为 `host:port`。这样：
- 避免与全局 `--host` 冲突
- 简化参数（一个参数代替两个）
- 符合 Go 标准库 `net.Listen` / `http.Server.Addr` 的惯例

## R4: 端点路径

**Decision**: 使用默认端点路径 `/mcp`

**Rationale**: 这是 mcp-go 的默认值，也是 MCP 规范推荐的标准路径。客户端连接地址为 `http://host:port/mcp`。
