# Data Model: MCP HTTP Transport

本特性不引入新的数据模型或持久化实体。所有数据流复用现有的 storage 层（reader.go）。

## 运行时实体

### MCP Server 实例
- **MCPServer**: 已有实例，注册 4 个工具，与传输层无关
- **StreamableHTTPServer**: mcp-go 提供的 HTTP 传输层包装，持有 MCPServer 引用

### 配置参数（运行时）

| 参数 | 来源 | 默认值 | 说明 |
|------|------|--------|------|
| httpMode | `--http` 标志 | false | 是否启用 HTTP 传输 |
| addr | `--addr` 参数 | `0.0.0.0:8080` | HTTP 监听地址 |
| dataDir | 全局 `--dir` 参数 | `{appDir}/data` | 数据目录（已有） |

## 数据流

```
MCP Client (HTTP)
    │
    ▼
StreamableHTTPServer (mcp-go)
    │  POST /mcp (JSON-RPC)
    ▼
MCPServer (mcp-go)
    │  Tool dispatch
    ▼
Tool Handlers (mcp/tools.go)
    │
    ▼
Storage Layer (storage/reader.go)
    │
    ▼
File System (data/{author}/...)
```
