# Implementation Plan: MCP HTTP Transport

**Branch**: `002-mcp-http-transport` | **Date**: 2026-02-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-mcp-http-transport/spec.md`

## Summary

为现有 MCP Server 添加 HTTP Streamable 传输模式，使其可以作为网络服务在 NAS 上长期运行。通过 mcp-go v0.43.2 的 `StreamableHTTPServer` 原生支持实现，仅需修改 2 个文件，不影响现有 stdio 模式和工具逻辑。

## Technical Context

**Language/Version**: Go 1.24
**Primary Dependencies**: mark3labs/mcp-go v0.43.2（`StreamableHTTPServer`）、urfave/cli/v3
**Storage**: 文件系统（只读，已有 storage/reader.go）
**Testing**: go test（手动集成测试 + curl 验证）
**Target Platform**: Linux (NAS/Raspberry Pi)、Tailscale 内网
**Project Type**: single（CLI 工具 + 可选 HTTP 服务）
**Performance Goals**: 单次工具调用 <5s 响应
**Constraints**: 无认证（Tailscale 内网）、优雅关闭 <10s
**Scale/Scope**: 单用户/少量并发

## Constitution Check

*项目未配置 constitution，跳过此检查。*

## Project Structure

### Documentation (this feature)

```text
specs/002-mcp-http-transport/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: API research
├── data-model.md        # Phase 1: Data model
├── quickstart.md        # Phase 1: Quick start guide
├── contracts/           # Phase 1: Endpoint contracts
│   └── mcp-http-endpoint.md
└── tasks.md             # Phase 2 output (via /speckit.tasks)
```

### Source Code (repository root)

```text
mcp/
├── server.go            # 修改：添加 ServeHTTP() 函数
├── tools.go             # 不变
└── search.go            # 不变

main.go                  # 修改：mcp 子命令添加 --http、--addr 标志
```

**Structure Decision**: 复用现有 `mcp/` 包结构，仅在 `server.go` 中新增 HTTP 服务函数。无需创建新文件或新包。

## Implementation Design

### 核心变更 1: `mcp/server.go` — 添加 ServeHTTP 函数

```go
// ServeHTTP 以 HTTP Streamable 传输模式启动 MCP Server
func ServeHTTP(s *server.MCPServer, addr string) error {
    httpServer := server.NewStreamableHTTPServer(s)

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        slog.Info("MCP HTTP Server 正在监听", "addr", addr, "endpoint", "/mcp")
        errCh <- httpServer.Start(addr)
    }()

    select {
    case err := <-errCh:
        return err  // 启动失败（如端口占用）
    case <-ctx.Done():
        slog.Info("正在关闭 MCP HTTP Server...")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        return httpServer.Shutdown(shutdownCtx)
    }
}
```

**关键设计**:
- `Start()` 在 goroutine 中运行（阻塞调用）
- `select` 同时监听启动错误和关闭信号
- 10 秒 shutdown 超时确保优雅关闭
- 日志输出实际监听地址

### 核心变更 2: `main.go` — MCP 子命令添加标志

```go
{
    Name:  "mcp",
    Usage: "以 MCP Server 模式启动，通过 stdio 或 HTTP 提供文档检索服务",
    Flags: []cli.Flag{
        &cli.BoolFlag{
            Name:  "http",
            Usage: "以 HTTP Streamable 模式启动（默认为 stdio 模式）",
        },
        &cli.StringFlag{
            Name:  "addr",
            Value: "0.0.0.0:8080",
            Usage: "HTTP 监听地址（格式: host:port）",
        },
    },
    Action: func(ctx context.Context, cmd *cli.Command) error {
        slog.Info("MCP Server 启动中", "dataDir", cfg.DataDir)
        s := mcpserver.NewServer(cfg.DataDir, version)
        if cmd.Bool("http") {
            return mcpserver.ServeHTTP(s, cmd.String("addr"))
        }
        slog.Info("MCP Server 已就绪，等待连接...")
        return mcpserver.Serve(s)
    },
}
```

**关键设计**:
- `--http` 布尔标志切换传输模式
- `--addr` 使用 `host:port` 格式（与 Go 标准库一致），避免与全局 `--host` 冲突
- 不设 `--http` 时行为完全不变（向后兼容）

## Complexity Tracking

无 constitution 违规，无需记录。
